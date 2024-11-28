package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

func sendRequests(url string, requestsPerGoroutine int, wg *sync.WaitGroup) {
	defer wg.Done()

	client := &http.Client{
		Timeout: 1 * time.Second, // Уменьшаем таймаут для быстрого завершения запросов
	}

	for i := 0; i < requestsPerGoroutine; i++ {
		resp, err := client.Get(url)
		if err != nil {
			fmt.Printf("Request failed: %v\n", err)
			continue
		}
		resp.Body.Close() // Ensure response body is closed
		fmt.Printf("Response code: %d\n", resp.StatusCode)

		// Добавляем короткую задержку между запросами, чтобы нагружать сервис не только количеством запросов
		time.Sleep(100 * time.Millisecond) // уменьшите для увеличения нагрузки
	}
}

func main() {
	url := "http://localhost:8081/" // Убедитесь, что ваш сервис работает на этом порту
	totalRequests := 50000          // Увеличиваем общее количество запросов
	concurrentGoroutines := 200     // Увеличиваем количество параллельных goroutines

	requestsPerGoroutine := totalRequests / concurrentGoroutines
	var wg sync.WaitGroup

	start := time.Now()

	// Создаем много параллельных goroutines
	for i := 0; i < concurrentGoroutines; i++ {
		wg.Add(1)
		go sendRequests(url, requestsPerGoroutine, &wg)
	}

	wg.Wait()

	duration := time.Since(start)
	fmt.Printf("Completed %d requests in %v\n", totalRequests, duration)
}
