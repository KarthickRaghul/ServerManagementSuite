package settings

import (
	generaldb "github.com/kishore-001/ServerManagementSuite/backend/db/gen/general"
	"encoding/json"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"strings"
)

// Standard response structures
type ErrorResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func HandleAddUser(queries *generaldb.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only allow POST
		if r.Method != http.MethodPost {
			sendError(w, "Only POST method allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse request body
		var req struct {
			Name     string `json:"username"`
			Role     string `json:"role"`
			Email    string `json:"email"`
			Password string `json:"password"` // Plain password from frontend
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			sendError(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Validate required fields
		if req.Name == "" || req.Role == "" || req.Email == "" || req.Password == "" {
			sendError(w, "All fields are required", http.StatusBadRequest)
			return
		}

		// Validate role
		if req.Role != "admin" && req.Role != "viewer" {
			sendError(w, "Role must be 'admin' or 'viewer'", http.StatusBadRequest)
			return
		}

		// Hash the password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			sendError(w, "Failed to hash password: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Create user in database
		newUser, err := queries.CreateUser(r.Context(), generaldb.CreateUserParams{
			Name:         strings.TrimSpace(req.Name),
			Role:         strings.TrimSpace(req.Role),
			Email:        strings.TrimSpace(req.Email),
			PasswordHash: string(hashedPassword), // Store hashed password
		})
		if err != nil {
			// Check for duplicate email/name errors
			if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
				sendError(w, "User with this email or name already exists", http.StatusConflict)
				return
			}
			sendError(w, "Failed to create user: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Prepare response (exclude password hash)
		response := map[string]interface{}{
			"status":  "success",
			"message": "User created successfully",
			"user": map[string]interface{}{
				"id":    newUser.ID,
				"name":  newUser.Name,
				"role":  newUser.Role,
				"email": newUser.Email,
			},
		}

		// Send successful response
		sendPostSuccess(w, response)
	}
}

// Standard response functions
func sendPostSuccess(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(data)
}

func sendError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	errorResp := ErrorResponse{
		Status:  "failed",
		Message: message,
	}
	json.NewEncoder(w).Encode(errorResp)
}
