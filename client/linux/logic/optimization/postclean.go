package optimization

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
)

type SuccessResponse struct {
	Status string `json:"status"`
}

func sendPostSuccess(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
	response := SuccessResponse{
		Status: "success",
	}
	json.NewEncoder(w).Encode(response)
}

// cleanDir tries to remove all files/subfolders in a directory (but not the dir itself)
func cleanDir(path string) error {
	dir, err := os.Open(path)
	if err != nil {
		return err
	}
	defer dir.Close()

	names, err := dir.Readdirnames(-1)
	if err != nil {
		return err
	}

	for _, name := range names {
		fullpath := filepath.Join(path, name)
		err = os.RemoveAll(fullpath)
		if err != nil {
			return err
		}
	}
	return nil
}

func HandleFileClean(w http.ResponseWriter, r *http.Request) {
	// Check for POST method
	if r.Method != http.MethodPost {
		sendError(w, "Only POST method allowed", http.StatusMethodNotAllowed)
		return
	}

	// Set content type
	w.Header().Set("Content-Type", "application/json")

	usr, err := user.Current()
	if err != nil {
		sendError(w, "Failed to get current user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	userCache := filepath.Join(usr.HomeDir, ".cache")

	dirs := []string{"/tmp", "/var/tmp", userCache}
	var cleaned []string
	var failed []string

	for _, dir := range dirs {
		err := cleanDir(dir)
		if err == nil {
			cleaned = append(cleaned, dir)
		} else {
			failed = append(failed, fmt.Sprintf("%s (%v)", dir, err))
		}
	}

	// Special handling for this function - allow "partial" status as exception
	if len(failed) > 0 && len(cleaned) > 0 {
		// Partial success - some directories cleaned, some failed
		resp := map[string]interface{}{
			"status":  "partial",
			"message": fmt.Sprintf("Cleaned: %v. Failed: %v", cleaned, failed),
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
		return
	} else if len(failed) > 0 && len(cleaned) == 0 {
		// Complete failure - no directories cleaned
		sendError(w, fmt.Sprintf("Failed to clean all directories: %v", failed), http.StatusInternalServerError)
		return
	}

	// Complete success - all directories cleaned
	sendPostSuccess(w)
}


