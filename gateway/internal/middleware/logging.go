package middleware

import (
	"math/rand"
	"time"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func Logger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// Create structured log entry
		logger := log.With().
			Str("method", param.Method).
			Str("path", param.Path).
			Int("status_code", param.StatusCode).
			Dur("latency", param.Latency).
			Str("client_ip", param.ClientIP).
			Str("user_agent", param.Request.UserAgent()).
			Logger()

		
		switch {
		case param.StatusCode >= 500:
			logger.Error().Msg("Server error")
		case param.StatusCode >= 400:
			logger.Warn().Msg("Client error")
		default:
			logger.Info().Msg("Request processed")
		}

		return "" 
	})
}


func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}

		
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)

		
		c.Request = c.Request.WithContext(
			log.With().Str("request_id", requestID).Logger().WithContext(c.Request.Context()),
		)

		c.Next()
	}
}


func generateRequestID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}


func randomString(n int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

