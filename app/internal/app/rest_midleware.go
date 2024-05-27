package app

import (
	"fmt"
	"s3MediaStreamer/app/internal/logs"
	"time"

	"github.com/gin-gonic/gin"
)

// LoggingMiddlewareFunc - тип для функции логирования.
type LoggingMiddlewareFunc func(format string, args ...interface{})

func formatCallerInfo(file string) string {
	return fmt.Sprintf("%s:", file)
}

func LoggingMiddlewareAdapter(logger *logs.Logger) LoggingMiddlewareFunc {
	return func(format string, args ...interface{}) {
		LogWithLogrusf(logger, format, args...)
	}
}

// LoggingMiddleware создает промежуточное ПО для логирования HTTP-запросов.
func LoggingMiddleware(logFunc LoggingMiddlewareFunc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Starting time request
		startTime := time.Now()

		// Processing request
		ctx.Next()

		// End Time request
		endTime := time.Now()

		// execution time
		latencyTime := endTime.Sub(startTime)

		// Request method
		reqMethod := ctx.Request.Method

		// Request route
		reqURI := ctx.Request.RequestURI

		// status code
		statusCode := ctx.Writer.Status()

		// Request IP
		clientIP := ctx.ClientIP()

		callerInfo := formatCallerInfo(ctx.HandlerName())

		logFunc("%s HTTP REQUEST - METHOD: %s, URI: %s, STATUS: %d, LATENCY: %s, CLIENT_IP: %s",
			callerInfo, reqMethod, reqURI, statusCode, latencyTime, clientIP)
	}
}
