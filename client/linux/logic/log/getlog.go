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
	// Parse from ISO format: "2025-06-10T08:00:05+05:30 ROG-G15 kernel: message"
	parts := strings.Fields(line)
	if len(parts) >= 3 {
		// Skip timestamp and hostname, get service name
		servicePart := parts[2]
		// Remove brackets and process info if present
		if strings.Contains(servicePart, "[") {
			servicePart = strings.Split(servicePart, "[")[0]
		}
		if strings.HasSuffix(servicePart, ":") {
			servicePart = strings.TrimSuffix(servicePart, ":")
		}
		return servicePart
	}
	return "unknown"
}

func HandleLog(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Support both GET and POST methods
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get number of lines from query parameter
	numLines := 100
	if n := r.URL.Query().Get("lines"); n != "" {
		if parsed, err := strconv.Atoi(n); err == nil && parsed > 0 {
			numLines = parsed
		}
	}

	log.Printf("🔍 [CLIENT-LOG] Method: %s, Lines: %d", r.Method, numLines)

	// Parse date/time filter from request body (for POST requests)
	var filter LogFilterRequest
	if r.Method == http.MethodPost && r.Body != nil {
		if err := json.NewDecoder(r.Body).Decode(&filter); err != nil {
			log.Printf("❌ [CLIENT-LOG] Failed to parse filter: %v", err)
			http.Error(w, "Invalid filter format", http.StatusBadRequest)
			return
		}
		log.Printf("🔍 [CLIENT-LOG] Received filter: %+v", filter)
	}

	// ✅ Build journalctl arguments based on your working command
	var args []string

	if filter.Date != "" {
		// ✅ Use the exact format that works
		since := filter.Date + " 00:00:00"
		until := filter.Date + " 23:59:59"

		if filter.Time != "" {
			// If time is specified, start from that time
			since = fmt.Sprintf("%s %s", filter.Date, filter.Time)
		}

		args = []string{
			"--since", since,
			"--until", until,
			"--no-pager",
			"--output=short-iso",
		}
		log.Printf("🔍 [CLIENT-LOG] Using date range: %s to %s", since, until)

	} else if filter.Time != "" {
		// Only time provided - use today's date with specified time
		today := time.Now().Format("2006-01-02")
		since := fmt.Sprintf("%s %s", today, filter.Time)
		args = []string{
			"-n", fmt.Sprintf("%d", numLines),
			"--since", since,
			"--no-pager",
			"--output=short-iso",
		}
		log.Printf("🔍 [CLIENT-LOG] Using time filter: %s", since)

	} else {
		// No date/time filter - get recent logs
		args = []string{
			"-n", fmt.Sprintf("%d", numLines),
			"--no-pager",
			"--output=short-iso",
		}
		log.Printf("🔍 [CLIENT-LOG] Using recent logs filter")
	}

	log.Printf("🔍 [CLIENT-LOG] Executing: journalctl %s", strings.Join(args, " "))

	// Execute journalctl command
	cmd := exec.Command("journalctl", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("❌ [CLIENT-LOG] journalctl error: %v, output: %s", err, string(output))
		http.Error(w, "Error fetching logs: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("✅ [CLIENT-LOG] journalctl executed successfully, output length: %d", len(output))

	// ✅ Parse journalctl output in ISO format
	lines := strings.Split(string(output), "\n")
	var entries []LogEntry

	for _, line := range lines {
		if len(line) < 20 {
			continue
		}

		// ✅ Parse ISO format: "2025-06-10T08:00:05+05:30 ROG-G15 kernel: message"
		// Extract timestamp (first 25 characters include timezone)
		if len(line) < 25 {
			continue
		}

		timestampStr := line[:25]            // "2025-06-10T08:00:05+05:30"
		rest := strings.TrimSpace(line[26:]) // Everything after timestamp

		// Parse timestamp
		t, err := time.Parse("2006-01-02T15:04:05-07:00", timestampStr)
		if err != nil {
			log.Printf("⚠️ [CLIENT-LOG] Failed to parse timestamp '%s': %v", timestampStr, err)
			t = time.Now()
		}

		// Parse the rest: "ROG-G15 kernel: message"
		msgParts := strings.SplitN(rest, ": ", 2)
		if len(msgParts) < 2 {
			continue
		}

		application := parseApplication(line)
		message := msgParts[1]
		level := parseLogLevel(message)

		entry := LogEntry{
			Timestamp:   t.Format(time.RFC3339),
			Level:       level,
			Application: application,
			Message:     message,
		}
		entries = append(entries, entry)
	}

	// ✅ Apply line limit if date filter was used and we have too many entries
	if filter.Date != "" && len(entries) > numLines {
		// Take the most recent entries
		entries = entries[len(entries)-numLines:]
	}

	log.Printf("✅ [CLIENT-LOG] Parsed %d log entries", len(entries))

	// Return JSON response
	if err := json.NewEncoder(w).Encode(entries); err != nil {
		log.Printf("❌ [CLIENT-LOG] Failed to encode response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
