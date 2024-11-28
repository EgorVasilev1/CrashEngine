package main

import (
	"html/template"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Define metrics
var (
	requestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"path", "method", "status"},
	)

	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Histogram of response time for handler in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"path", "method", "status"},
	)

	serviceStatus = struct {
		sync.RWMutex
		IsUp bool
	}{IsUp: true}
)

func init() {
	// Register metrics
	prometheus.MustRegister(requestsTotal)
	prometheus.MustRegister(requestDuration)
}

// Middleware to collect metrics
func metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(rw, r)

		duration := time.Since(start).Seconds()
		path := r.URL.Path
		method := r.Method
		status := rw.statusCode

		// Update metrics
		requestsTotal.WithLabelValues(path, method, http.StatusText(status)).Inc()
		requestDuration.WithLabelValues(path, method, http.StatusText(status)).Observe(duration)

		// Simulate a failure if latency exceeds 1 second
		if duration > 1 {
			serviceStatus.Lock()
			serviceStatus.IsUp = false
			serviceStatus.Unlock()
		}
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello, World!"))
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	serviceStatus.RLock()
	isUp := serviceStatus.IsUp
	serviceStatus.RUnlock()

	status := "Service is UP"
	if !isUp {
		status = "Service is DOWN"
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(status))
}

// HTML visualization handler
func visualizationHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := `
	<!DOCTYPE html>
	<html>
	<head>
		<title>Service Metrics</title>
		<script>
			function reload() {
				fetch('/metrics').then(res => res.text()).then(data => {
					document.getElementById('metrics').innerText = data;
				});
				setTimeout(reload, 5000);
			}
			window.onload = reload;
		</script>
	</head>
	<body>
		<h1>Service Metrics</h1>
		<pre id="metrics">Loading metrics...</pre>
	</body>
	</html>
	`
	t, _ := template.New("metrics").Parse(tmpl)
	t.Execute(w, nil)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", helloHandler)
	mux.HandleFunc("/status", statusHandler)
	mux.HandleFunc("/visualization", visualizationHandler)

	// Add Prometheus metrics endpoint
	mux.Handle("/metrics", promhttp.Handler())

	// Wrap with metrics middleware
	handler := metricsMiddleware(mux)

	log.Println("Starting server on :8081")
	if err := http.ListenAndServe(":8081", handler); err != nil {
		log.Fatalf("could not start server: %v\n", err)
	}
}
