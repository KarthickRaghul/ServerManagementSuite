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

func parseApplication(line string) string {
	parts := strings.Fields(line)
	if len(parts) >= 3 {
		return parts[2]
	}
	return "unknown"
}

func HandleLog(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	numLines := 100
	if n := r.URL.Query().Get("lines"); n != "" {
		if parsed, err := strconv.Atoi(n); err == nil && parsed > 0 {
			numLines = parsed
		}
	}

	// Parse optional filter from request body
	var filter LogFilterRequest
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&filter)
	}

	// Build journalctl arguments
	args := []string{"-n", fmt.Sprintf("%d", numLines), "--no-pager", "--output=short-iso"}

	// If date or time filtering provided, add --since flag
	if filter.Date != "" || filter.Time != "" {
		since := ""
		if filter.Date != "" && filter.Time != "" {
			// Both date and time provided
			since = fmt.Sprintf("%s %s", filter.Date, filter.Time)
		} else if filter.Date != "" {
			// Only date provided
			since = filter.Date
		} else if filter.Time != "" {
			// Only time provided, use today as date
			since = time.Now().Format("2006-01-02") + " " + filter.Time
		}
		if since != "" {
			args = append(args, "--since", since)
		}
	}

	cmd := exec.Command("journalctl", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		http.Error(w, "Error fetching logs: "+err.Error(), http.StatusInternalServerError)
		return
	}

	lines := strings.Split(string(output), "\n")
	var entries []LogEntry

	for _, line := range lines {
		if len(line) < 20 {
			continue
		}

		// Format: "2025-06-11T09:45:36+0530 ROG-G15 kitty[1234]: message..."
		timestamp := line[:19]
		rest := strings.TrimSpace(line[20:])

		msgParts := strings.SplitN(rest, ": ", 2)
		if len(msgParts) < 2 {
			continue
		}

		application := parseApplication(rest)
		message := msgParts[1]
		level := parseLogLevel(message)

		// Parse and reformat timestamp to UTC ISO format
		t, err := time.Parse("2006-01-02T15:04:05", timestamp)
		if err != nil {
			t = time.Now()
		}

		entry := LogEntry{
			Timestamp:   t.Format(time.RFC3339),
			Level:       level,
			Application: application,
			Message:     message,
		}
		entries = append(entries, entry)
	}

	json.NewEncoder(w).Encode(entries)
}

