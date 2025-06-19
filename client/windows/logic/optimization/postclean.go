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

// HandleFileClean handles POST requests to clean temporary/cache folders on Windows
func HandleFileClean(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	// Optionally check user (but not used)
	_, err := os.UserHomeDir()
	if err != nil {
		http.Error(w, "Failed to get user home directory", http.StatusInternalServerError)
		return
	}

	// Windows common temp/cache directories
	tempDir := os.TempDir()                         // e.g., C:\Users\You\AppData\Local\Temp
	localAppData := os.Getenv("LOCALAPPDATA")       // e.g., C:\Users\You\AppData\Local
	userTemp := filepath.Join(localAppData, "Temp") // Redundant but kept for completeness

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

	status := "success"
	message := fmt.Sprintf("Cleaned: %v", cleaned)
	if len(failed) > 0 {
		status = "partial"
		message = fmt.Sprintf("Cleaned: %v. Failed: %v", cleaned, failed)
	}

	resp := map[string]interface{}{
		"status":  status,
		"message": message,
	}

	json.NewEncoder(w).Encode(resp)
}
