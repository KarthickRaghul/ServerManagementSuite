package optimization

import (
	"bufio"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/process"
)

type ServiceInfo struct {
	PID     int32  `json:"pid"`
	User    string `json:"user"`
	Name    string `json:"name"`
	Cmdline string `json:"cmdline"`
	Type    string `json:"type"` // "user" only (no system services)
}

type ServiceListResult struct {
	Status    string        `json:"status"`
	Message   string        `json:"message"`
	Services  []ServiceInfo `json:"services"`
	Timestamp string        `json:"timestamp"`
}

// getUserServices retrieves user services for all non-root users
func getUserServices() (map[string]string, error) {
	userServices := make(map[string]string) // service name -> username

	// Get all regular users (non-root, UID >= 1000)
	users, err := getAllRegularUsers()
	if err != nil {
		return userServices, err
	}

	for _, username := range users {
		// Skip root user explicitly
		if username == "root" {
			continue
		}

		// Try to get user services using systemctl --user
		cmd := exec.Command("systemctl", "--user", "-M", username+"@", "list-units", "--type=service", "--state=running", "--no-pager", "--no-legend")
		out, err := cmd.Output()
		if err != nil {
			// Fallback: try alternative method
			cmd = exec.Command("sudo", "-u", username, "XDG_RUNTIME_DIR=/run/user/$(id -u "+username+")", "systemctl", "--user", "list-units", "--type=service", "--state=running", "--no-pager", "--no-legend")
			out, err = cmd.Output()
			if err != nil {
				continue // Skip this user if we can't get their services
			}
		}

		scanner := bufio.NewScanner(strings.NewReader(string(out)))
		for scanner.Scan() {
			line := scanner.Text()
			fields := strings.Fields(line)
			if len(fields) > 0 {
				unitName := fields[0]
				if strings.HasSuffix(unitName, ".service") {
					service := strings.TrimSuffix(unitName, ".service")
					userServices[service] = username
				}
			}
		}
	}

	return userServices, nil
}

// getAllRegularUsers returns all regular users (UID >= 1000, excluding root)
func getAllRegularUsers() ([]string, error) {
	var users []string

	// Read /etc/passwd to get all users
	cmd := exec.Command("getent", "passwd")
	out, err := cmd.Output()
	if err != nil {
		return users, err
	}

	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Split(line, ":")
		if len(fields) >= 3 {
			username := fields[0]
			uidStr := fields[2]

			uid, err := strconv.Atoi(uidStr)
			if err != nil {
				continue
			}

			// Include only regular users (UID >= 1000) and explicitly exclude root and system users
			if uid >= 1000 && uid != 65534 && username != "root" {
				users = append(users, username)
			}
		}
	}

	return users, nil
}

func HandleListService(w http.ResponseWriter, r *http.Request) {
	// Check for GET method
	if r.Method != http.MethodGet {
		sendError(w, "Only GET method allowed", http.StatusMethodNotAllowed)
		return
	}

	// Set content type
	w.Header().Set("Content-Type", "application/json")

	// Get only user services (no system services)
	userServices, err := getUserServices()
	if err != nil {
		// Don't fail completely if user services can't be retrieved
		userServices = make(map[string]string)
	}

	// Get all processes
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

		// Skip root processes entirely
		if username == "root" || username == "" {
			continue
		}

		nameLower := strings.ToLower(name)

		// Check if this is a user service
		if userServiceOwner, exists := userServices[nameLower]; exists {
			svc := ServiceInfo{
				PID:     proc.Pid,
				User:    userServiceOwner,
				Name:    name,
				Cmdline: cmdline,
				Type:    "user",
			}
			services = append(services, svc)
		}

		// Also include user processes that look like services (non-root only)
		if isUserService(name, cmdline) {
			svc := ServiceInfo{
				PID:     proc.Pid,
				User:    username,
				Name:    name,
				Cmdline: cmdline,
				Type:    "user",
			}
			services = append(services, svc)
		}
	}

	result := ServiceListResult{
		Status:    "success",
		Message:   "List of user services (excluding root)",
		Services:  services,
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
	}

	// Send successful GET response with data
	sendGetSuccess(w, result)
}


// isUserService checks if a process looks like a user service
func isUserService(name, cmdline string) bool {
	nameLower := strings.ToLower(name)
	cmdlineLower := strings.ToLower(cmdline)

	// Common user service patterns
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

	// Check for daemon-like names
	if strings.HasSuffix(nameLower, "d") && len(nameLower) > 2 {
		return true
	}

	// Check for service-like command lines
	if strings.Contains(cmdlineLower, "--daemon") ||
		strings.Contains(cmdlineLower, "--service") ||
		strings.Contains(cmdlineLower, "serve") {
		return true
	}

	return false
}
