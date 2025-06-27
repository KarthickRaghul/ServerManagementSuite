package optimization

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

// Standard response structures matching other handlers
type SuccessResponse struct {
	Status string `json:"status"`
}

type PartialResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// Standard success response function
func sendPostSuccess(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
	response := SuccessResponse{
		Status: "success",
	}
	json.NewEncoder(w).Encode(response)
}

// Partial success response function
func sendPartialResponse(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
	response := PartialResponse{
		Status: "partial",
	}
	json.NewEncoder(w).Encode(response)
}

type CleanResult struct {
	Cleaned []string
	Failed  []string
}

func cleanDir(path string) CleanResult {
	result := CleanResult{}

	entries, err := os.ReadDir(path)
	if err != nil {
		result.Failed = append(result.Failed, fmt.Sprintf("%s (read error: %v)", path, err))
		return result
	}

	for _, entry := range entries {
		fullpath := filepath.Join(path, entry.Name())
		err := os.RemoveAll(fullpath)
		if err != nil {
			result.Failed = append(result.Failed, fmt.Sprintf("%s (%v)", fullpath, err))
		} else {
			result.Cleaned = append(result.Cleaned, fullpath)
		}
	}

	return result
}

func HandleFileClean(w http.ResponseWriter, r *http.Request) {
	// Check for POST method
	if r.Method != http.MethodPost {
		sendError(w, "Only POST method allowed", http.StatusMethodNotAllowed)
		return
	}

	// Set content type
	w.Header().Set("Content-Type", "application/json")

	// Windows temp directories
	tempDir := os.TempDir()
	localAppData := os.Getenv("LOCALAPPDATA")
	userTemp := filepath.Join(localAppData, "Temp")

	dirs := []string{tempDir, userTemp}
	var totalCleaned, totalFailed []string

	for _, dir := range dirs {
		result := cleanDir(dir)
		totalCleaned = append(totalCleaned, result.Cleaned...)
		totalFailed = append(totalFailed, result.Failed...)
	}

	// ✅ Updated response logic with standard format
	switch {
	case len(totalFailed) == 0:
		// Complete success - all files cleaned
		sendPostSuccess(w)
	case len(totalCleaned) > 0:
		// Partial success - some files cleaned, some failed
		sendPartialResponse(w)
	default:
		// Complete failure - no files cleaned
		sendError(w, "Failed to clean any files. All directories may be protected or in use", http.StatusInternalServerError)
	}
}
