package optimization

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

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

	// Windows common temp/cache directories
	tempDir := os.TempDir()                   // e.g., C:\Users\You\AppData\Local\Temp
	localAppData := os.Getenv("LOCALAPPDATA") // e.g., C:\Users\You\AppData\Local
	userTemp := filepath.Join(localAppData, "Temp")

	dirs := []string{tempDir, userTemp}
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

	// Handle partial success
	if len(failed) > 0 && len(cleaned) > 0 {
		resp := map[string]interface{}{
			"status":  "partial",
			"message": fmt.Sprintf("Cleaned: %v. Failed: %v", cleaned, failed),
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
		return
	} 

	// Handle complete failure
	if len(failed) > 0 && len(cleaned) == 0 {
		sendError(w, fmt.Sprintf("Failed to clean all directories: %v", failed), http.StatusInternalServerError)
		return
	}

	// Complete success
	sendPostSuccess(w)
}


func sendPostSuccess(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
	response := map[string]string{
		"status": "success",
	}
	json.NewEncoder(w).Encode(response)
}
