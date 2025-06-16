package settings

import (
	"backend/config"
	generaldb "backend/db/gen/general"
	"encoding/json"
	"net/http"
)

func HandleRemoveUser(queries *generaldb.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check admin authorization
		user, ok := config.GetUserFromContext(r)
		if !ok {
			http.Error(w, "User context not found", http.StatusInternalServerError)
			return
		}

		// Parse request body
		var req struct {
			Name string `json:"username"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Validate input
		if req.Name == "" {
			http.Error(w, "Username is required", http.StatusBadRequest)
			return
		}

		// Prevent admin from deleting themselves
		if req.Name == user.Username {
			http.Error(w, "Cannot delete your own account", http.StatusBadRequest)
			return
		}

		// Check if user exists before deletion
		_, err := queries.GetUserByName(r.Context(), req.Name)
		if err != nil {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}

		// Delete the user (correct parameter - just the name string)
		err = queries.DeleteUserByName(r.Context(), req.Name)
		if err != nil {
			http.Error(w, "Failed to remove user", http.StatusInternalServerError)
			return
		}

		// Success response
		response := map[string]interface{}{
			"status":       "success",
			"message":      "User removed successfully",
			"deleted_user": req.Name,
			"deleted_by":   user.Username,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}
