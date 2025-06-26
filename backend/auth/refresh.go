package auth

import (
	db "github.com/kishore-001/ServerManagementSuite/backend/db/gen/general"
	"context"
	"encoding/json"
	"net/http"
)

func HandleRefresh(dbQueries *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Get refresh token from cookie
		cookie, err := r.Cookie("refresh_token")
		if err != nil {
			http.Error(w, "No refresh token", http.StatusUnauthorized)
			return
		}

		// Validate refresh token in database
		session, err := dbQueries.GetRefreshToken(context.Background(), cookie.Value)
		if err != nil {
			http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
			return
		}

		// Get user details for new access token
		user, err := dbQueries.GetUserByName(context.Background(), session.Username)
		if err != nil {
			http.Error(w, "User not found", http.StatusUnauthorized)
			return
		}

		// Generate new access token
		accessToken, err := GenerateAccessToken(user.Name, user.Role)
		if err != nil {
			http.Error(w, "Failed to generate token", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":       "ok",
			"access_token": accessToken,
		})
	}
}
