package settings

import (
	"github.com/kishore-001/ServerManagementSuite/backend/config"
	generaldb "github.com/kishore-001/ServerManagementSuite/backend/db/gen/general"
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
)

func HandleRemoveUser(queries *generaldb.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only allow POST/DELETE method
		if r.Method != http.MethodPost && r.Method != http.MethodDelete {
			sendError(w, "Only POST or DELETE method allowed", http.StatusMethodNotAllowed)
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

		// Parse request body
		var req struct {
			Name string `json:"username"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			sendError(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Validate input
		if req.Name == "" {
			sendError(w, "Username is required", http.StatusBadRequest)
			return
		}

		// Trim whitespace
		req.Name = strings.TrimSpace(req.Name)

		// Prevent admin from deleting themselves
		if req.Name == user.Username {
			sendError(w, "Cannot delete your own account", http.StatusBadRequest)
			return
		}

		// Check if user exists before deletion
		_, err := queries.GetUserByName(r.Context(), req.Name)
		if err == sql.ErrNoRows {
			sendError(w, "User not found", http.StatusNotFound)
			return
		} else if err != nil {
			sendError(w, "Database error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Delete the user
		err = queries.DeleteUserByName(r.Context(), req.Name)
		if err != nil {
			sendError(w, "Failed to remove user: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Success response
		response := map[string]interface{}{
			"status":       "success",
			"message":      "User removed successfully",
			"deleted_user": req.Name,
			"deleted_by":   user.Username,
		}

		// Send successful response
		sendGetSuccess(w, response)
	}
}
