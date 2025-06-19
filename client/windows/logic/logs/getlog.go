package log

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type LogEntry struct {
	Timestamp   string `json:"timestamp"`
	Level       string `json:"level"`
	Application string `json:"application"`
	Message     string `json:"message"`
}

type LogFilterRequest struct {
	Date string `json:"date"` // Format: "YYYY-MM-DD"
	Time string `json:"time"` // Format: "HH:MM:SS"
}

// parseLogLevel - very basic level parsing based on keywords
func parseLogLevel(message string) string {
	lower := strings.ToLower(message)
	switch {
	case strings.Contains(lower, "error"):
		return "error"
	case strings.Contains(lower, "warn"):
		return "warning"
	default:
		return "info"
	}
}

func HandleLog(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	numLines := 100
	if n := r.URL.Query().Get("lines"); n != "" {
		if parsed, err := strconv.Atoi(n); err == nil && parsed > 0 {
			numLines = parsed
		}
	}

	var filter LogFilterRequest
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&filter)
	}

	// Construct PowerShell command
	sinceFilter := ""
	if filter.Date != "" || filter.Time != "" {
		timestamp := ""
		if filter.Date != "" && filter.Time != "" {
			timestamp = fmt.Sprintf("%s %s", filter.Date, filter.Time)
		} else if filter.Date != "" {
			timestamp = filter.Date
		} else if filter.Time != "" {
			timestamp = time.Now().Format("2006-01-02") + " " + filter.Time
		}

		sinceFilter = fmt.Sprintf(`| Where-Object {$_.TimeCreated -ge (Get-Date "%s")}`, timestamp)
	}

	powershellCmd := fmt.Sprintf(`Get-WinEvent -LogName System -MaxEvents %d %s | Select-Object TimeCreated, Message, ProviderName | Format-List`, numLines, sinceFilter)
	cmd := exec.Command("powershell", "-Command", powershellCmd)

	output, err := cmd.CombinedOutput()
	if err != nil {
		http.Error(w, "Error fetching logs: "+err.Error(), http.StatusInternalServerError)
		return
	}

	entries := []LogEntry{}
	lines := strings.Split(string(output), "\n")

	var currentEntry LogEntry
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "TimeCreated") {
			timeStr := strings.TrimPrefix(line, "TimeCreated : ")
			t, err := time.Parse(time.RFC3339, timeStr)
			if err == nil {
				currentEntry.Timestamp = t.Format(time.RFC3339)
			} else {
				currentEntry.Timestamp = time.Now().Format(time.RFC3339)
			}
		} else if strings.HasPrefix(line, "ProviderName") {
			currentEntry.Application = strings.TrimPrefix(line, "ProviderName : ")
		} else if strings.HasPrefix(line, "Message") {
			currentEntry.Message = strings.TrimPrefix(line, "Message : ")
			currentEntry.Level = parseLogLevel(currentEntry.Message)
		} else if line == "" && currentEntry.Message != "" {
			entries = append(entries, currentEntry)
			currentEntry = LogEntry{}
		}
	}

	json.NewEncoder(w).Encode(entries)
}
