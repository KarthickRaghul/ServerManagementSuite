package settings

import (
	generaldb "backend/db/gen/general"
	"encoding/json"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

func HandleAddUser(queries *generaldb.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse request body
		var req struct {
			Name     string `json:"username"`
			Role     string `json:"role"`
			Email    string `json:"email"`
			Password string `json:"password"` // Plain password from frontend
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Validate required fields
		if req.Name == "" || req.Role == "" || req.Email == "" || req.Password == "" {
			http.Error(w, "All fields are required", http.StatusBadRequest)
			return
		}

		// Hash the password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Failed to hash password", http.StatusInternalServerError)
			return
		}

		// Create user in database
		newUser, err := queries.CreateUser(r.Context(), generaldb.CreateUserParams{
			Name:         req.Name,
			Role:         req.Role,
			Email:        req.Email,
			PasswordHash: string(hashedPassword), // Store hashed password
		})
		if err != nil {
			// Check for duplicate email/name errors
			if err.Error() == "pq: duplicate key value violates unique constraint" {
				http.Error(w, "User with this email or name already exists", http.StatusConflict)
				return
			}
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
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

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	}
}
