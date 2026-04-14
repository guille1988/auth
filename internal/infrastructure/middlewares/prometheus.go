package middlewares

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var httpRequestsTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests",
	},
	[]string{"method", "path", "status"},
)

func Prometheus() gin.HandlerFunc {
	return func(context *gin.Context) {
		context.Next()
		httpRequestsTotal.WithLabelValues(
			context.Request.Method,
			context.FullPath(),
			strconv.Itoa(context.Writer.Status()),
		).Inc()
	}
}
