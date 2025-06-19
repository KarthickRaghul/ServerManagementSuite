package optimization

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
)

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
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		http.Error(w, "Only GET allowed", http.StatusMethodNotAllowed)
		return
	}

	if _, err := user.Current(); err != nil {
		http.Error(w, "Failed to get current user", http.StatusInternalServerError)
		return
	}

	// Windows common temp/cache directories
	userTemp := filepath.Join(os.Getenv("TEMP"))
	userAppData := filepath.Join(os.Getenv("APPDATA"))
	localAppData := filepath.Join(os.Getenv("LOCALAPPDATA"))

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

	json.NewEncoder(w).Encode(resp)
}
