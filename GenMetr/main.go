package main

import (
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)
	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Histogram of response time for handler in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)
	httpErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_errors_total",
			Help: "Total number of HTTP errors",
		},
		[]string{"method", "path", "status"},
	)
)

func init() {
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpRequestDuration)
	prometheus.MustRegister(httpErrorsTotal)
}

func main() {
	router := gin.Default()
	router.Use(prometheusMiddleware())

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	router.GET("/generate", generateRequests)

	// Запуск сервера
	err := router.Run("0.0.0.0:8080")
	if err != nil {
		return
	}
}

func prometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start).Seconds()
		status := c.Writer.Status()

		httpRequestsTotal.WithLabelValues(c.Request.Method, c.FullPath(), http.StatusText(status)).Inc()
		httpRequestDuration.WithLabelValues(c.Request.Method, c.FullPath()).Observe(duration)

		if status >= 400 {
			httpErrorsTotal.WithLabelValues(c.Request.Method, c.FullPath(), http.StatusText(status)).Inc()
		}
	}
}

func generateRequests(c *gin.Context) {
	// Список возможных статус-кодов
	statuses := []int{200, 201, 400, 401, 403, 404, 500, 502, 503}
	randomStatus := statuses[rand.Intn(len(statuses))]

	// Генерация случайной задержки
	time.Sleep(time.Millisecond * time.Duration(rand.Intn(500)))

	// Возврат случайного статуса
	if randomStatus >= 400 {
		c.JSON(randomStatus, gin.H{"error": http.StatusText(randomStatus)})
	} else {
		c.JSON(randomStatus, gin.H{"message": "OK"})
	}
}
