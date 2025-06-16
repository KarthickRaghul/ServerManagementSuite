package optimization

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
)

// sizeOfDir recursively calculates the total size of files in a directory
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

	usr, _ := user.Current()
	userCache := filepath.Join(usr.HomeDir, ".cache")

	dirs := []string{"/tmp", "/var/tmp", userCache}
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
