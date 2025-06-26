package auth

import (
	db "github.com/kishore-001/ServerManagementSuite/backend/db/gen/general"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"
)

type TokenVerifyRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type TokenVerifyResponse struct {
	Status string `json:"status"`
}

func HandleVerify(dbQueries *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Get refresh token from cookie
		cookie, err := r.Cookie("refresh_token")
		if err != nil {
			// No refresh token cookie found
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(TokenVerifyResponse{Status: "unauthorized"})
			return
		}

		refreshToken := cookie.Value
		if refreshToken == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(TokenVerifyResponse{Status: "unauthorized"})
			return
		}

		// Check if refresh token exists in database
		session, err := dbQueries.GetRefreshToken(context.Background(), refreshToken)
		if err == sql.ErrNoRows {
			// Token not found in database
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(TokenVerifyResponse{Status: "unauthorized"})
			return
		} else if err != nil {
			// Database error
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(TokenVerifyResponse{Status: "unauthorized"})
			return
		}

		// Check if token is expired
		if session.ExpiresAt.Before(time.Now()) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(TokenVerifyResponse{Status: "unauthorized"})
			return
		}

		// Token is valid and not expired
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(TokenVerifyResponse{Status: "authorized"})
	}
}
