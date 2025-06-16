package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"os"
	"strings"
)

const tokenFilePath = "./auth/token.hash"

func TokenAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Unauthorized Access", http.StatusUnauthorized)
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		tokenHash := sha256.Sum256([]byte(token))
		tokenHex := hex.EncodeToString(tokenHash[:])

		storedHash, err := os.ReadFile(tokenFilePath)
		if err != nil {
			http.Error(w, "Server error: token file not found", http.StatusInternalServerError)
			return
		}

		if strings.TrimSpace(tokenHex) != strings.TrimSpace(string(storedHash)) {
			http.Error(w, "Unauthorized Access", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
