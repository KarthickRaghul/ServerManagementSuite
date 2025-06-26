package settings

import (
	"github.com/kishore-001/ServerManagementSuite/backend/config"
	generaldb "github.com/kishore-001/ServerManagementSuite/backend/db/gen/general"
	"encoding/json"
	"net/http"
)

func HandleListUsers(queries *generaldb.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check if it's a GET request
		if r.Method != http.MethodGet {
			sendError(w, "Only GET method allowed", http.StatusMethodNotAllowed)
			return
		}

		// Check admin authorization
		user, ok := config.GetUserFromContext(r)
		if !ok {
			sendError(w, "User context not found", http.StatusInternalServerError)
			return
		}

		if user.Role != "admin" {
			sendError(w, "Admin access required", http.StatusForbidden)
			return
		}

		// Get all users from database
		users, err := queries.ListUsers(r.Context())
		if err != nil {
			sendError(w, "Failed to fetch users: "+err.Error(), http.StatusInternalServerError)
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

		// Send successful response
		sendGetSuccess(w, response)
	}
}

// sendGetSuccess sends successful GET response with data
func sendGetSuccess(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)
}
