package main

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte("vanakamdamapla") // 🔐 Make sure to keep this secret!

func main() {
	// Set claims
	claims := jwt.MapClaims{
		"sub": "192.168.1.101",                  // Subject: IP or device ID
		"iat": time.Now().Unix(),                // Issued at
		"exp": time.Now().Add(time.Hour).Unix(), // Expires in 1 hour
		"iss": "snsms-backend",                  // Issuer
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		panic(err)
	}

	fmt.Println("Generated JWT token:")
	fmt.Println(tokenString)
}
