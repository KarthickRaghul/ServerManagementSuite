package optimization

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
)

// Standard response structures
type ErrorResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// sizeOfDir calculates total size of all files in the directory (recursively)
func sizeOfDir(path string) (int64, error) {
	var total int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			total += info.Size()
		}
		return nil
	})
	return total, err
}

// HandleFileInfo returns info about the directories to be cleaned and their current sizes
func HandleFileInfo(w http.ResponseWriter, r *http.Request) {
	// Check for GET method
	if r.Method != http.MethodGet {
		sendError(w, "Only GET method allowed", http.StatusMethodNotAllowed)
		return
	}

	// Set content type
	w.Header().Set("Content-Type", "application/json")

	_, err := user.Current()
	if err != nil {
		sendError(w, "Failed to get current user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Windows common temp/cache directories
	userTemp := filepath.Clean(os.Getenv("TEMP"))
	userAppData := filepath.Clean(os.Getenv("APPDATA"))
	localAppData := filepath.Clean(os.Getenv("LOCALAPPDATA"))

	dirs := []string{
		userTemp,
		filepath.Join(userAppData, "Temp"),
		filepath.Join(localAppData, "Temp"),
	}

	sizes := make(map[string]int64)
	var failed []string

	for _, dir := range dirs {
		size, err := sizeOfDir(dir)
		if err == nil {
			sizes[dir] = size
		} else {
			failed = append(failed, fmt.Sprintf("%s (%v)", dir, err))
		}
	}

	resp := map[string]interface{}{
		"folders": dirs,
		"sizes":   sizes,
		"failed":  failed,
	}

	// Send successful GET response with data
	sendGetSuccess(w, resp)
}

// sendGetSuccess sends successful GET response with data
func sendGetSuccess(w http.ResponseWriter, data interface{}) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)
}

// sendError sends standardized error response
func sendError(w http.ResponseWriter, message string, statusCode int) {
	w.WriteHeader(statusCode)
	errorResp := ErrorResponse{
		Status:  "failed",
		Message: message,
	}
	json.NewEncoder(w).Encode(errorResp)
}
