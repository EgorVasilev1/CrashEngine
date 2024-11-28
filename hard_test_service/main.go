package main

import (
	"bytes"
	"fmt"
	"net/http"
	"sync"
	"time"
)

func sendRequests(url string, requestsPerGoroutine int, wg *sync.WaitGroup) {
	defer wg.Done()

	client := &http.Client{
		Timeout: 0, // Бесконечный таймаут для создания постоянной нагрузки
	}

	// Огромные тела запросов (для POST-запросов)
	body := bytes.Repeat([]byte("A"), 10*1024*1024) // 10 МБ данных в каждом запросе

	for i := 0; i < requestsPerGoroutine; i++ {
		// Бесконечно долгие POST-запросы
		resp, err := client.Post(url, "application/octet-stream", bytes.NewReader(body))
		if err != nil {
			fmt.Printf("Request failed: %v\n", err)
			continue
		}
		resp.Body.Close() // Ensure response body is closed
		fmt.Printf("Response code: %d\n", resp.StatusCode)

		// Небольшая задержка между запросами для увеличения нагрузки
		time.Sleep(50 * time.Millisecond) // Уменьшаем задержку для создания еще более высокой нагрузки
	}
}

func main() {
	url := "http://localhost:8081/" // Убедитесь, что ваш сервис работает на этом порту
	totalRequests := 100000         // Увеличиваем количество запросов
	concurrentGoroutines := 500     // Увеличиваем количество параллельных goroutines

	requestsPerGoroutine := totalRequests / concurrentGoroutines
	var wg sync.WaitGroup

	start := time.Now()

	// Создаем огромное количество параллельных goroutines с большим объемом данных в каждом запросе
	for i := 0; i < concurrentGoroutines; i++ {
		wg.Add(1)
		go sendRequests(url, requestsPerGoroutine, &wg)
	}

	wg.Wait()

	duration := time.Since(start)
	fmt.Printf("Completed %d requests in %v\n", totalRequests, duration)
}
