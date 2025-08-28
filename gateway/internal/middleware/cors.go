package middleware

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)


func CORS() gin.HandlerFunc {
	config := cors.DefaultConfig()
	
	
	config.AllowAllOrigins = true
	
	
	config.AllowHeaders = []string{
		"Origin",
		"Content-Type",
		"Accept",
		"Authorization",
		"X-Requested-With",
		"X-Gateway",
	}
	
	
	config.AllowMethods = []string{
		"GET",
		"POST",
		"PUT",
		"DELETE",
		"OPTIONS",
		"PATCH",
	}
	
	
	config.AllowCredentials = true
	
	config.ExposeHeaders = []string{
		"Content-Length",
		"X-Gateway",
		"X-Forwarded-By",
	}
	
	return cors.New(config)
}




