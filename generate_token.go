package main

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func main() {
	// JWT secret from environment
	secretKey := "smartseller_super_secret_session_key_min_32_chars"

	// Admin user details (from database query)
	userID := "3c3ee730-335a-4c0f-82fb-5b07d0a606a0"
	email := "test@example.com"
	name := "Test User"

	// Create claims for the JWT (24 hours expiry)
	expiryTime := time.Now().Add(24 * time.Hour)
	claims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"name":    name,
		"role":    "admin",
		"exp":     expiryTime.Unix(),
		"iat":     time.Now().Unix(),
	}

	// Create JWT token with claims
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the secret key
	tokenString, err := jwtToken.SignedString([]byte(secretKey))
	if err != nil {
		fmt.Printf("Error signing JWT token: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Generated JWT Token:\n%s\n", tokenString)
}