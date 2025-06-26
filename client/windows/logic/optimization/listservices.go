package optimization

import (
	"net/http"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/process"
)

type ServiceInfo struct {
	PID     int32  `json:"pid"`
	User    string `json:"user"`
	Name    string `json:"name"`
	Cmdline string `json:"cmdline"`
	Type    string `json:"type"` // "user"
}

type ServiceListResult struct {
	Status    string        `json:"status"`
	Message   string        `json:"message"`
	Services  []ServiceInfo `json:"services"`
	Timestamp string        `json:"timestamp"`
}


// getUserServices returns an empty map (Windows doesn't have systemd)
func getUserServices() (map[string]string, error) {
	return map[string]string{}, nil
}


func HandleListService(w http.ResponseWriter, r *http.Request) {
	// Check for GET method
	if r.Method != http.MethodGet {
		sendError(w, "Only GET method allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// Get user services (empty on Windows)
	userServices, _ := getUserServices()

	procs, err := process.Processes()
	if err != nil {
		sendError(w, "Failed to fetch processes: "+err.Error(), http.StatusInternalServerError)
		return
	}

	services := []ServiceInfo{}

	for _, proc := range procs {
		name, _ := proc.Name()
		cmdline, _ := proc.Cmdline()
		username, _ := proc.Username()

		// Skip SYSTEM processes
		if strings.Contains(strings.ToLower(username), "system") || username == "" {
			continue
		}

		nameLower := strings.ToLower(name)

		if userServiceOwner, exists := userServices[nameLower]; exists {
			services = append(services, ServiceInfo{
				PID:     proc.Pid,
				User:    userServiceOwner,
				Name:    name,
				Cmdline: cmdline,
				Type:    "user",
			})
		}

		// Heuristic check for service-like processes
		if isUserService(name, cmdline) {
			services = append(services, ServiceInfo{
				PID:     proc.Pid,
				User:    username,
				Name:    name,
				Cmdline: cmdline,
				Type:    "user",
			})
		}
	}

	result := ServiceListResult{
		Status:    "success",
		Message:   "List of user services (excluding SYSTEM)",
		Services:  services,
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
	}

	sendGetSuccess(w, result)
}

func isUserService(name, cmdline string) bool {
	nameLower := strings.ToLower(name)
	cmdlineLower := strings.ToLower(cmdline)

	userServicePatterns := []string{
		"node", "python", "java", "php", "ruby", "go", "npm", "yarn",
		"docker", "podman", "code", "electron", "chrome", "firefox",
		"discord", "slack", "telegram", "steam", "spotify",
		"server", "daemon", "service", "worker", "agent",
	}

	for _, pattern := range userServicePatterns {
		if strings.Contains(nameLower, pattern) || strings.Contains(cmdlineLower, pattern) {
			return true
		}
	}

	if strings.HasSuffix(nameLower, "d") && len(nameLower) > 2 {
		return true
	}

	if strings.Contains(cmdlineLower, "--daemon") ||
		strings.Contains(cmdlineLower, "--service") ||
		strings.Contains(cmdlineLower, "serve") {
		return true
	}

	return false
}

