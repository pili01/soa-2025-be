package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Services ServicesConfig
	Auth     AuthConfig
}

type ServerConfig struct {
	Port         string
	ReadTimeout  int
	WriteTimeout int
}

type ServicesConfig struct {
	Blog         string
	Image        string
	Stakeholders string
	Tours        string
	ToursAPI     string
	Follower     string
	Purchase     string
}

type AuthConfig struct {
	JWTSecret string
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("../config")

	// Set defaults
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.readTimeout", 30)
	viper.SetDefault("server.writeTimeout", 30)

	// Environment variables override
	viper.AutomaticEnv()

	// Try to read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	config := &Config{
		Server: ServerConfig{
			Port:         getEnv("GATEWAY_PORT", "8080"),
			ReadTimeout:  getEnvAsInt("GATEWAY_READ_TIMEOUT", 30),
			WriteTimeout: getEnvAsInt("GATEWAY_WRITE_TIMEOUT", 30),
		},
		Services: ServicesConfig{
			Blog:         getEnv("BLOG_SERVICE_URL", "http://blog-service:3000"),
			Image:        getEnv("IMAGE_SERVICE_URL", "http://image-service:3001"),
			Stakeholders: getEnv("STAKEHOLDERS_SERVICE_URL", "http://stakeholders-service:8081"),
			Follower:     getEnv("FOLLOWER_SERVICE_URL", "http://follower-service:8083"),
			Tours:        getEnv("TOURS_SERVICE_GRPC_URL", "tours-service:50051"),
			ToursAPI:     getEnv("TOURS_SERVICE_API_URL", "http://tours-service:8081"),
			Purchase:     getEnv("PURCHASE_SERVICE_URL", "http://purchase-service:8080"),
		},
		Auth: AuthConfig{
			JWTSecret: getEnv("JWT_SECRET", "your-secret-key"),
		},
	}

	return config, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
