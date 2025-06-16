package settings

import (
	"backend/config"
	generaldb "backend/db/gen/general"
	"encoding/json"
	"net/http"
)

func HandleListUsers(queries *generaldb.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check if it's a GET request
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Check admin authorization
		user, ok := config.GetUserFromContext(r)
		if !ok {
			http.Error(w, "User context not found", http.StatusInternalServerError)
			return
		}

		if user.Role != "admin" {
			http.Error(w, "Admin access required", http.StatusForbidden)
			return
		}

		// Get all users from database
		users, err := queries.ListUsers(r.Context())
		if err != nil {
			http.Error(w, "Failed to fetch users", http.StatusInternalServerError)
			return
		}

		// Prepare response (exclude password hashes)
		var userList []map[string]interface{}
		for _, u := range users {
			userList = append(userList, map[string]interface{}{
				"id":    u.ID,
				"name":  u.Name,
				"role":  u.Role,
				"email": u.Email,
			})
		}

		response := map[string]interface{}{
			"status": "success",
			"users":  userList,
			"count":  len(userList),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}
