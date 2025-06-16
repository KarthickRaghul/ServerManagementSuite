package auth

import (
	db "backend/db/gen/general"
	"context"
	"database/sql"
	"encoding/json"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"time"
)

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	Status      string `json:"status"`
	AccessToken string `json:"access_token,omitempty"`
	Message     string `json:"message,omitempty"` // Optional for errors
}

func writeJSON(w http.ResponseWriter, status int, response loginResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(response)
}

func HandleLogin(dbQueries *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, loginResponse{Status: "error", Message: "Method not allowed"})
			return
		}

		var req loginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, loginResponse{Status: "error", Message: "Invalid request payload"})
			return
		}

		user, err := dbQueries.GetUserByName(context.Background(), req.Username)
		if err == sql.ErrNoRows {
			writeJSON(w, http.StatusUnauthorized, loginResponse{Status: "error", Message: "Invalid credentials"})
			return
		} else if err != nil {
			writeJSON(w, http.StatusInternalServerError, loginResponse{Status: "error", Message: "Database error"})
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
			writeJSON(w, http.StatusUnauthorized, loginResponse{Status: "error", Message: "Invalid credentials"})
			return
		}

		accessToken, err := GenerateAccessToken(req.Username, user.Role)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, loginResponse{Status: "error", Message: "Failed to generate access token"})
			return
		}

		var refreshToken string
		existingSession, err := dbQueries.GetValidRefreshTokenByUser(context.Background(), req.Username)

		if err == sql.ErrNoRows {
			refreshToken, err = GenerateRefreshToken()
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, loginResponse{Status: "error", Message: "Failed to generate refresh token"})
				return
			}

			err = dbQueries.SaveRefreshToken(context.Background(), db.SaveRefreshTokenParams{
				Username:     req.Username,
				RefreshToken: refreshToken,
				ExpiresAt:    time.Now().Add(7 * 24 * time.Hour),
			})
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, loginResponse{Status: "error", Message: "Failed to save session"})
				return
			}
		} else if err != nil {
			writeJSON(w, http.StatusInternalServerError, loginResponse{Status: "error", Message: "Database error"})
			return
		} else {
			refreshToken = existingSession.RefreshToken
		}

		// Set refresh token in HttpOnly cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "refresh_token",
			Value:    refreshToken,
			Path:     "/",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
			MaxAge:   7 * 24 * 60 * 60,
		})

		// Final successful login response
		writeJSON(w, http.StatusOK, loginResponse{
			Status:      "ok",
			AccessToken: accessToken,
		})
	}
}
