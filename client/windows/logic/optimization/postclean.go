package optimization

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"time"
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

	var finalErr error
	for _, name := range names {
		fullpath := filepath.Join(path, name)
		err = os.RemoveAll(fullpath)
		if err != nil {
			fmt.Printf("Failed to remove %s: %v\n", fullpath, err)
			if finalErr == nil {
				finalErr = fmt.Errorf("%s (%v)", fullpath, err)
			} else {
				finalErr = fmt.Errorf("%v; %s (%v)", finalErr, fullpath, err)
			}
		}
	}
	return finalErr
}

func HandleFileClean(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	usr, err := user.Current()
	if err != nil {
		http.Error(w, "Cannot get current user", http.StatusInternalServerError)
		return
	}

	userTemp := os.Getenv("TEMP")
	if userTemp == "" {
		userTemp = os.Getenv("TMP")
	}
	if userTemp == "" {
		userTemp = filepath.Join(usr.HomeDir, "AppData", "Local", "Temp")
	}

	windowsTemp := filepath.Join(os.Getenv("SystemRoot"), "Temp")
	if windowsTemp == "" {
		windowsTemp = `C:\Windows\Temp`
	}

	userCache := filepath.Join(usr.HomeDir, "AppData", "Local", "Cache")

	dirs := []string{userTemp, windowsTemp, userCache}

	var cleaned []string
	var failed []string

	for _, dir := range dirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue
		}

		err := cleanDir(dir)
		if err == nil {
			cleaned = append(cleaned, dir)
		} else {
			failed = append(failed, fmt.Sprintf("%s (%v)", dir, err))
		}
	}

	status := "success"
	if len(failed) > 0 {
		status = "partial"
	}

	resp := map[string]interface{}{
		"status":    status,
		"cleaned":   cleaned,
		"failed":    failed,
		"timestamp": time.Now().Format("2006-01-02 15:04:05"),
	}

	json.NewEncoder(w).Encode(resp)
}
