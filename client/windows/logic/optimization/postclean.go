package optimization

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

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
	if r.Method != http.MethodPost {
		sendError(w, "Only POST method allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

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

	switch {
	case len(totalFailed) == 0:
		// All cleaned
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "success",
		})
	case len(totalCleaned) > 0:
		// Partial
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "partial",
			"message": fmt.Sprintf("The Data are Cleaned Partial because file are open"),
		})
	default:
		// Full failure
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "failed",
			"message": fmt.Sprintf("Failed to clean any files: %v", totalFailed),
		})
	}
}
