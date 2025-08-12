package utils

import (
	"encoding/base64"
	"fmt"
	"time"
	"strconv"
	"os"
	"github.com/dgrijalva/jwt-go"
)

var jwtKey []byte

func init() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		panic("JWT_SECRET is not set in environment")
	}

	key, err := base64.StdEncoding.DecodeString(secret)
	if err != nil {
		panic(fmt.Errorf("failed to decode JWT_SECRET: %w", err))
	}
	jwtKey = key
}

type Claims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.StandardClaims
}

func GenerateToken(username, role string) (string, error) {
	expirationStr := os.Getenv("JWT_EXPIRATION")
	expirationMinutes, err := strconv.Atoi(expirationStr)
	expirationTime := time.Now().Add(time.Duration(expirationMinutes) * time.Minute)

	claims := &Claims{
		Username: username,
		Role:     role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

func ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}
