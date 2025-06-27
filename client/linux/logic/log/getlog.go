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

type ErrorResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
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

// validateDateTime validates and parses date/time strings
func validateDateTime(dateStr, timeStr string) (time.Time, time.Time, bool, error) {
	now := time.Now()

	// Default to today if no date provided
	targetDate := dateStr
	if targetDate == "" {
		targetDate = now.Format("2006-01-02")
	}

	// Validate date format
	parsedDate, err := time.Parse("2006-01-02", targetDate)
	if err != nil {
		return time.Time{}, time.Time{}, false, fmt.Errorf("invalid date format. Expected YYYY-MM-DD, got: %s", targetDate)
	}

	// Default to start of day if no time provided
	targetTime := timeStr
	if targetTime == "" {
		targetTime = "00:00:00"
	}

	// Validate time format
	_, err = time.Parse("15:04:05", targetTime)
	if err != nil {
		return time.Time{}, time.Time{}, false, fmt.Errorf("invalid time format. Expected HH:MM:SS, got: %s", targetTime)
	}

	// Create start time
	startTime := time.Date(parsedDate.Year(), parsedDate.Month(), parsedDate.Day(),
		0, 0, 0, 0, time.Local)

	if timeStr != "" {
		// Parse the specific time
		timeParts := strings.Split(targetTime, ":")
		hour, _ := strconv.Atoi(timeParts[0])
		minute, _ := strconv.Atoi(timeParts[1])
		second, _ := strconv.Atoi(timeParts[2])

		startTime = time.Date(parsedDate.Year(), parsedDate.Month(), parsedDate.Day(),
			hour, minute, second, 0, time.Local)
	}

	// End time logic
	var endTime time.Time
	var useEndTime bool

	if timeStr == "" {
		// If only date provided, use end of day to limit to that specific date
		endTime = time.Date(parsedDate.Year(), parsedDate.Month(), parsedDate.Day(),
			23, 59, 59, 0, time.Local)
		useEndTime = true
	} else {
		// If specific time provided, add 30 minutes for the time window
		endTime = startTime.Add(30 * time.Minute)
		useEndTime = true
	}

	return startTime, endTime, useEndTime, nil
}

func HandleLog(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Default number of lines
	numLines := 100
	if n := r.URL.Query().Get("lines"); n != "" {
		if parsed, err := strconv.Atoi(n); err == nil && parsed > 0 && parsed <= 10000 {
			numLines = parsed
		}
	}

	log.Printf("🔍 [CLIENT-LOG] Method: %s, Lines: %d", r.Method, numLines)

	var filter LogFilterRequest
	if r.Method == http.MethodPost && r.Body != nil {
		if err := json.NewDecoder(r.Body).Decode(&filter); err != nil {
			log.Printf("❌ [CLIENT-LOG] Failed to parse filter: %v", err)
			sendError(w, "Invalid filter format", http.StatusBadRequest)
			return
		}
		log.Printf("🔍 [CLIENT-LOG] Received filter: %+v", filter)
	}

	args := []string{"--no-pager", "--output=json"}

	// Handle date/time filtering
	if filter.Date != "" || filter.Time != "" {
		startTime, endTime, _, err := validateDateTime(filter.Date, filter.Time)
		if err != nil {
			log.Printf("❌ [CLIENT-LOG] DateTime validation error: %v", err)
			sendError(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Format start and end time for journalctl
		since := startTime.Format("2006-01-02 15:04:05")
		until := endTime.Format("2006-01-02 15:04:05")
		args = append(args, "--since", since, "--until", until)

		if filter.Time != "" {
			log.Printf("🔍 [CLIENT-LOG] Using 30-minute time window: %s to %s (will limit to first %d logs)", since, until, numLines)
		} else {
			log.Printf("🔍 [CLIENT-LOG] Using date range filter: %s to %s (will limit to %d logs)", since, until, numLines)
		}
	} else {
		log.Printf("🔍 [CLIENT-LOG] No filter provided. Using recent logs only")
		// When no filter is provided, get logs from last 24 hours
		since := time.Now().Add(-24 * time.Hour).Format("2006-01-02 15:04:05")
		args = append(args, "--since", since)
	}

	// Always get logs in chronological order (oldest first) when using time filters
	// This ensures we get the FIRST N logs from the time window, not the last N
	if filter.Date != "" || filter.Time != "" {
		args = append(args, "--reverse")
		log.Printf("🔍 [CLIENT-LOG] Getting logs in chronological order from time window")
	} else {
		// For no filter, get most recent logs (newest first)
		args = append(args, "-n", fmt.Sprintf("%d", numLines))
		log.Printf("🔍 [CLIENT-LOG] Getting last %d recent logs", numLines)
	}

	log.Printf("🔍 [CLIENT-LOG] Executing: journalctl %s", strings.Join(args, " "))

	cmd := exec.Command("journalctl", args...)
	output, err := cmd.Output()
	if err != nil {
		log.Printf("❌ [CLIENT-LOG] journalctl error: %v", err)

		// Check if it's a journalctl-specific error
		if exitError, ok := err.(*exec.ExitError); ok {
			stderr := string(exitError.Stderr)
			if strings.Contains(stderr, "Invalid time") || strings.Contains(stderr, "time") {
				sendError(w, "Invalid date/time format for system logs", http.StatusBadRequest)
				return
			}
		}

		sendError(w, "Error fetching logs: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if len(output) == 0 {
		sendError(w, "No logs found for the specified criteria", http.StatusNotFound)
		return
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var entries []LogEntry

	// Process logs and limit to requested number
	processedCount := 0
	for i, line := range lines {
		// Stop when we reach the requested number of entries
		if processedCount >= numLines {
			break
		}

		if strings.TrimSpace(line) == "" {
			continue
		}

		var raw map[string]interface{}
		if err := json.Unmarshal([]byte(line), &raw); err != nil {
			log.Printf("⚠️ [CLIENT-LOG] Failed to parse log line %d: %v", i, err)
			continue
		}

		// Extract timestamp
		tsStr, ok := raw["__REALTIME_TIMESTAMP"].(string)
		if !ok {
			continue
		}

		tsInt, err := strconv.ParseInt(tsStr, 10, 64)
		if err != nil {
			continue
		}

		timestamp := time.UnixMicro(tsInt).Format(time.RFC3339)

		// Extract message
		msg, _ := raw["MESSAGE"].(string)
		if msg == "" {
			continue // Skip entries without messages
		}

		// Extract application name
		app, _ := raw["SYSLOG_IDENTIFIER"].(string)
		if app == "" {
			app, _ = raw["_COMM"].(string)
		}
		if app == "" {
			app, _ = raw["_SYSTEMD_UNIT"].(string)
		}
		if app == "" {
			app = "unknown"
		}

		level := parseLogLevel(msg)

		entries = append(entries, LogEntry{
			Timestamp:   timestamp,
			Level:       level,
			Application: app,
			Message:     msg,
		})

		processedCount++
	}

	if len(entries) == 0 {
		sendError(w, "No valid log entries found", http.StatusNotFound)
		return
	}

	log.Printf("✅ [CLIENT-LOG] Parsed %d log entries from time window", len(entries))

	if err := json.NewEncoder(w).Encode(entries); err != nil {
		log.Printf("❌ [CLIENT-LOG] Failed to encode response: %v", err)
		sendError(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func sendError(w http.ResponseWriter, message string, statusCode int) {
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{
		Status:  "failed",
		Message: message,
	})
}

