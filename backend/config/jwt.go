package config

import (
	"github.com/kishore-001/ServerManagementSuite/backend/auth"
	"context"
	"net/http"
	"strings"
)

type contextKey string

const UserContextKey contextKey = "user"

type UserInfo struct {
	Username string `json:"username"`
	Role     string `json:"role"`
}

// JWTMiddleware validates access tokens on protected routes
func JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing authorization header", http.StatusUnauthorized)
			return
		}

		// Extract token from "Bearer <token>"
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			http.Error(w, "Invalid authorization format", http.StatusUnauthorized)
			return
		}

		// Validate the access token
		claims, err := auth.ValidateAccessToken(tokenParts[1])
		if err != nil {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		// Add user info to request context
		userInfo := UserInfo{
			Username: claims.Username,
			Role:     claims.Role,
		}

		ctx := context.WithValue(r.Context(), UserContextKey, userInfo)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserFromContext extracts user info from request context
func GetUserFromContext(r *http.Request) (*UserInfo, bool) {
	user, ok := r.Context().Value(UserContextKey).(UserInfo)
	return &user, ok
}

// AdminOnly middleware - only allows admin users
func AdminOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := GetUserFromContext(r)
		if !ok {
			http.Error(w, "User context not found", http.StatusInternalServerError)
			return
		}

		if user.Role != "admin" {
			http.Error(w, "Admin access required", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
