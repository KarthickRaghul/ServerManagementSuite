package log

import (
	"encoding/json"
	"fmt"
	"log"
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

	// ✅ Support both GET and POST methods
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// ✅ Get number of lines from query parameter
	numLines := 100
	if n := r.URL.Query().Get("lines"); n != "" {
		if parsed, err := strconv.Atoi(n); err == nil && parsed > 0 {
			numLines = parsed
		}
	}

	log.Printf("🔍 [CLIENT-LOG] Method: %s, Lines: %d", r.Method, numLines)

	// ✅ Parse date/time filter from request body (for POST requests)
	var filter LogFilterRequest
	if r.Method == http.MethodPost && r.Body != nil {
		if err := json.NewDecoder(r.Body).Decode(&filter); err != nil {
			log.Printf("❌ [CLIENT-LOG] Failed to parse filter: %v", err)
			http.Error(w, "Invalid filter format", http.StatusBadRequest)
			return
		}
		log.Printf("🔍 [CLIENT-LOG] Received filter: %+v", filter)
	}

	// ✅ Build journalctl arguments
	args := []string{"-n", fmt.Sprintf("%d", numLines), "--no-pager", "--output=short-iso"}

	// ✅ Apply date/time filtering with better logic
	if filter.Date != "" || filter.Time != "" {
		since := ""

		if filter.Date != "" && filter.Time != "" {
			// Both date and time provided
			since = fmt.Sprintf("%s %s", filter.Date, filter.Time)
			log.Printf("🔍 [CLIENT-LOG] Using date+time filter: %s", since)
		} else if filter.Date != "" {
			// Only date provided - get logs from start of that day
			since = filter.Date + " 00:00:00"
			log.Printf("🔍 [CLIENT-LOG] Using date filter: %s", since)
		} else if filter.Time != "" {
			// Only time provided - use today's date with specified time
			today := time.Now().Format("2006-01-02")
			since = fmt.Sprintf("%s %s", today, filter.Time)
			log.Printf("🔍 [CLIENT-LOG] Using time filter: %s", since)
		}

		if since != "" {
			args = append(args, "--since", since)
			log.Printf("🔍 [CLIENT-LOG] journalctl args: %v", args)
		}
	}

	// ✅ Execute journalctl command
	cmd := exec.Command("journalctl", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("❌ [CLIENT-LOG] journalctl error: %v, output: %s", err, string(output))
		http.Error(w, "Error fetching logs: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("✅ [CLIENT-LOG] journalctl executed successfully, output length: %d", len(output))

	// ✅ Parse journalctl output
	lines := strings.Split(string(output), "\n")
	var entries []LogEntry

	for _, line := range lines {
		if len(line) < 20 {
			continue
		}

		// Format: "2025-06-19T20:45:36+0530 hostname service[pid]: message..."
		timestamp := line[:19]
		rest := strings.TrimSpace(line[20:])

		msgParts := strings.SplitN(rest, ": ", 2)
		if len(msgParts) < 2 {
			continue
		}

		application := parseApplication(rest)
		message := msgParts[1]
		level := parseLogLevel(message)

		// ✅ Parse and reformat timestamp to RFC3339
		t, err := time.Parse("2006-01-02T15:04:05", timestamp)
		if err != nil {
			// Fallback to current time if parsing fails
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

	log.Printf("✅ [CLIENT-LOG] Parsed %d log entries", len(entries))

	// ✅ Return JSON response
	if err := json.NewEncoder(w).Encode(entries); err != nil {
		log.Printf("❌ [CLIENT-LOG] Failed to encode response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

