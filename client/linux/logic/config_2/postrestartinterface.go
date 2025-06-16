package config_2

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

// HandleRestartInterfaces handles the request to restart all enabled interfaces
func HandleRestartInterfaces(w http.ResponseWriter, r *http.Request) {
	// Set content type
	w.Header().Set("Content-Type", "application/json")

	// Check for POST method
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get current time in UTC format
	currentTime := "2025-05-30 15:18:19" // Using provided UTC timestamp
	currentUser := "kishore-001"         // Using provided user login

	// Get all network interfaces
	interfaces, err := net.Interfaces()
	if err != nil {
		sendError(w, "Failed to get network interfaces", err)
		return
	}

	// List of interfaces that were restarted
	restartedInterfaces := []string{}

	// Try to restart each non-loopback, active interface
	for _, iface := range interfaces {
		// Skip loopback and virtual interfaces
		if iface.Flags&net.FlagLoopback != 0 ||
			strings.Contains(iface.Name, "docker") ||
			strings.Contains(iface.Name, "veth") ||
			strings.Contains(iface.Name, "br-") {
			continue
		}

		// Only restart if the interface is up
		if iface.Flags&net.FlagUp != 0 {
			err := restartInterface(iface.Name)
			if err != nil {
				fmt.Printf("Failed to restart interface %s: %v\n", iface.Name, err)
				continue
			}

			restartedInterfaces = append(restartedInterfaces, iface.Name)
		}
	}

	// Prepare response
	response := map[string]interface{}{
		"status":     "success",
		"message":    fmt.Sprintf("Restarted %d interfaces", len(restartedInterfaces)),
		"interfaces": restartedInterfaces,
		"timestamp":  currentTime,
		"user":       currentUser,
	}

	// If no interfaces were restarted, modify the response
	if len(restartedInterfaces) == 0 {
		response["status"] = "warning"
		response["message"] = "No interfaces were restarted"
	}

	// Send the response
	fmt.Printf("%d interfaces restarted by user %s\n", len(restartedInterfaces), currentUser)
	json.NewEncoder(w).Encode(response)
}

// restartInterface brings an interface down and then up again
func restartInterface(interfaceName string) error {
	// First, bring the interface down
	downCmd := exec.Command("ip", "link", "set", "dev", interfaceName, "down")
	output, err := downCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to bring interface down: %v, output: %s", err, string(output))
	}

	// Wait briefly to ensure the interface is fully down
	time.Sleep(500 * time.Millisecond)

	// Then, bring the interface up
	upCmd := exec.Command("ip", "link", "set", "dev", interfaceName, "up")
	output, err = upCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to bring interface up: %v, output: %s", err, string(output))
	}

	// Log the successful restart
	fmt.Printf("Successfully restarted interface %s\n", interfaceName)
	return nil
}
