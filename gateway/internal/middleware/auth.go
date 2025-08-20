package middleware

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog/log"
)


type JWTConfig struct {
	Secret string
}


func JWTAuth(config JWTConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if isPublicEndpoint(c.Request.URL.Path) {
			c.Next()
			return
		}

		
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			log.Warn().Str("path", c.Request.URL.Path).Msg("Missing Authorization header")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			log.Warn().Str("path", c.Request.URL.Path).Msg("Invalid Authorization header format")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization header format"})
			c.Abort()
			return
		}

		tokenString := tokenParts[1]

		
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			
			
			key, err := base64.StdEncoding.DecodeString(config.Secret)
			if err != nil {
				return nil, fmt.Errorf("failed to decode JWT secret: %w", err)
			}
			return key, nil
		})

		if err != nil {
			log.Warn().Err(err).Str("path", c.Request.URL.Path).Msg("JWT token validation failed")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		if !token.Valid {
			log.Warn().Str("path", c.Request.URL.Path).Msg("JWT token is invalid")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}
		
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			log.Warn().Str("path", c.Request.URL.Path).Msg("Invalid token claims format")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		c.Set("user_id", fmt.Sprintf("%v", claims["id"]))
		c.Set("username", fmt.Sprintf("%v", claims["username"]))
		c.Set("role", fmt.Sprintf("%v", claims["role"]))
		
		log.Debug().
			Str("user_id", fmt.Sprintf("%v", claims["id"])).
			Str("username", fmt.Sprintf("%v", claims["username"])).
			Str("path", c.Request.URL.Path).
			Msg("User authenticated")

		c.Next()
	}
}


func isPublicEndpoint(path string) bool {
	publicPaths := []string{
		"/health",
		"/metrics",
		"/api/register",
		"/api/login",
	}
	
	for _, publicPath := range publicPaths {
		if strings.HasPrefix(path, publicPath) {
			return true
		}
	}
	return false
}




