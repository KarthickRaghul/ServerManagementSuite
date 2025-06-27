package log

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"regexp"
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

func HandleLog(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	numLines := 100
	if n := r.URL.Query().Get("lines"); n != "" {
		if parsed, err := strconv.Atoi(n); err == nil && parsed > 0 && parsed <= 10000 {
			numLines = parsed
		}
	}

	var filter LogFilterRequest
	if r.Method == http.MethodPost && r.Body != nil {
		if err := json.NewDecoder(r.Body).Decode(&filter); err != nil {
			sendError(w, "Invalid filter format", http.StatusBadRequest)
			return
		}
	}

	log.Printf("Received filter: date=%s, time=%s", filter.Date, filter.Time)

	startTime, endTime, _, err := validateDateTime(filter.Date, filter.Time)
	if err != nil {
		sendError(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Printf("StartTime: %s, EndTime: %s", startTime.Format(time.RFC3339), endTime.Format(time.RFC3339))

	powershell := "powershell"
	var cmdStr string
	if filter.Date != "" || filter.Time != "" {
		cmdStr = fmt.Sprintf(`$since = Get-Date -Date '%s'; $until = Get-Date -Date '%s'; Get-WinEvent -FilterHashtable @{LogName='System'; StartTime=$since; EndTime=$until} -ErrorAction Stop | Select-Object -First %d -Property TimeCreated, Id, LevelDisplayName, ProviderName, Message | ConvertTo-Json -Compress`, startTime.Format(time.RFC3339), endTime.Format(time.RFC3339), numLines)
	} else {
		cmdStr = fmt.Sprintf(`Get-WinEvent -LogName System -MaxEvents %d -ErrorAction Stop | Select-Object -First %d -Property TimeCreated, Id, LevelDisplayName, ProviderName, Message | ConvertTo-Json -Compress`, numLines, numLines)
	}
	log.Printf("Generated PowerShell command: %s", cmdStr)

	cmd := exec.Command(powershell, "-Command", cmdStr)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		log.Printf("❌ PowerShell command error: %v\nOutput: %s", err, out.String())
		if strings.Contains(out.String(), "NoMatchingEventsFound") {
			sendError(w, "No logs found for the specified time or date filter", http.StatusNotFound)
			return
		}
		sendError(w, "Error fetching Windows logs", http.StatusInternalServerError)
		return
	}

	rawJSON := out.String()
	log.Printf("PowerShell output: %s", rawJSON)
	if strings.TrimSpace(rawJSON) == "" || strings.HasPrefix(rawJSON, "[]") {
		sendError(w, "No logs found for the specified time or date filter", http.StatusNotFound)
		return
	}

	var rawLogs []map[string]interface{}
	if err := json.Unmarshal([]byte(rawJSON), &rawLogs); err != nil {
		// Handle single object instead of array
		var single map[string]interface{}
		if err := json.Unmarshal([]byte(rawJSON), &single); err != nil {
			log.Printf("❌ JSON unmarshal error: %v", err)
			sendError(w, "Failed to parse Windows log output", http.StatusInternalServerError)
			return
		}
		rawLogs = append(rawLogs, single)
	}

	var entries []LogEntry
	for _, entry := range rawLogs {
		timeRaw, ok := entry["TimeCreated"]
		if !ok {
			log.Printf("❌ Missing TimeCreated in entry: %v", entry)
			continue
		}

		var timestamp string
		switch t := timeRaw.(type) {
		case string:
			// Handle Microsoft JSON date format: /Date(1234567890)/
			if strings.HasPrefix(t, "/Date(") && strings.HasSuffix(t, ")/") {
				re := regexp.MustCompile(`\d+`)
				millisStr := re.FindString(t)
				millis, err := strconv.ParseInt(millisStr, 10, 64)
				if err != nil {
					log.Printf("❌ Failed to parse TimeCreated milliseconds: %s, error: %v", t, err)
					continue
				}
				timestamp = time.UnixMilli(millis).Format(time.RFC3339)
			} else {
				// Try RFC3339
				parsed, err := time.Parse(time.RFC3339, t)
				if err != nil {
					log.Printf("❌ Failed to parse TimeCreated as RFC3339: %s, error: %v", t, err)
					continue
				}
				timestamp = parsed.Format(time.RFC3339)
			}
		case float64:
			timestamp = time.UnixMilli(int64(t)).Format(time.RFC3339)
		default:
			log.Printf("❌ Unexpected TimeCreated type: %T, value: %v", t, t)
			continue
		}

		msg, _ := entry["Message"].(string)
		if msg == "" {
			log.Printf("❌ Empty message in entry: %v", entry)
			continue
		}
		app, _ := entry["ProviderName"].(string)

		levelRaw, _ := entry["LevelDisplayName"].(string)
		level := strings.ToLower(levelRaw)

		switch level {
		case "information":
			level = "info"
		case "warning":
			level = "warning"
		case "error":
			level = "error"
		case "":
			level = "info"
		default:
			// optionally log or normalize others
			log.Printf("⚠️ Unknown log level: %s", levelRaw)
			level = "info"
		}

		if app == "" {
			app = "unknown"
		}

		entries = append(entries, LogEntry{
			Timestamp:   timestamp,
			Level:       strings.ToLower(level),
			Application: app,
			Message:     msg,
		})
		if len(entries) >= numLines {
			break
		}
	}

	if len(entries) == 0 {
		sendError(w, "No logs found for the specified time or date filter", http.StatusNotFound)
		return
	}

	if err := json.NewEncoder(w).Encode(entries); err != nil {
		log.Printf("❌ Failed to encode response: %v", err)
		sendError(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func validateDateTime(dateStr, timeStr string) (time.Time, time.Time, bool, error) {
	// Use local timezone (IST) for parsing
	localLoc, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		return time.Time{}, time.Time{}, false, fmt.Errorf("failed to load local timezone: %v", err)
	}

	now := time.Now().In(localLoc)
	if dateStr == "" {
		dateStr = now.Format("2006-01-02")
	}

	log.Printf("Parsing dateStr: %s", dateStr)
	parsedDate, err := time.ParseInLocation("2006-01-02", dateStr, localLoc)
	if err != nil {
		return time.Time{}, time.Time{}, false, fmt.Errorf("invalid date format. Expected YYYY-MM-DD")
	}

	if timeStr == "" {
		startTime := time.Date(parsedDate.Year(), parsedDate.Month(), parsedDate.Day(), 0, 0, 0, 0, localLoc)
		endTime := startTime.Add(24*time.Hour - time.Second)
		return startTime, endTime, true, nil
	}

	log.Printf("Parsing timeStr: %s", timeStr)
	parsedTime, err := time.Parse("15:04:05", timeStr)
	if err != nil {
		return time.Time{}, time.Time{}, false, fmt.Errorf("invalid time format. Expected HH:MM:SS")
	}

	startTime := time.Date(parsedDate.Year(), parsedDate.Month(), parsedDate.Day(), parsedTime.Hour(), parsedTime.Minute(), parsedTime.Second(), 0, localLoc)
	endTime := startTime.Add(30 * time.Minute)
	return startTime, endTime, true, nil
}

func sendError(w http.ResponseWriter, message string, statusCode int) {
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(ErrorResponse{
		Status:  "failed",
		Message: message,
	})
}
