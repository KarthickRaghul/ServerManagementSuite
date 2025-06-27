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

// HandleRestartInterfaces handles the request to restart all enabled interfaces (Windows version)
func HandleRestartInterfaces(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	currentTime := "2025-05-30 15:18:19"
	currentUser := "kishore-001"

	interfaces, err := net.Interfaces()
	if err != nil {
		sendError(w, "Failed to get network interfaces", http.StatusInternalServerError)
		return
	}

	restartedInterfaces := []string{}

	for _, iface := range interfaces {
		// Skip loopback and virtual interfaces
		if iface.Flags&net.FlagLoopback != 0 ||
			strings.Contains(iface.Name, "Loopback") ||
			strings.Contains(strings.ToLower(iface.Name), "vmware") ||
			strings.Contains(strings.ToLower(iface.Name), "virtual") {
			continue
		}

		// Only attempt restart if the interface is up
		if iface.Flags&net.FlagUp != 0 {
			err := restartInterfaceWindows(iface.Name)
			if err != nil {
				fmt.Printf("Failed to restart interface %s: %v\n", iface.Name, err)
				continue
			}
			restartedInterfaces = append(restartedInterfaces, iface.Name)
		}
	}

	response := map[string]interface{}{
		"status":     "success",
		"message":    fmt.Sprintf("Restarted %d interfaces", len(restartedInterfaces)),
		"interfaces": restartedInterfaces,
		"timestamp":  currentTime,
		"user":       currentUser,
	}

	if len(restartedInterfaces) == 0 {
		response["status"] = "warning"
		response["message"] = "No interfaces were restarted"
	}

	fmt.Printf("%d interfaces restarted by user %s\n", len(restartedInterfaces), currentUser)
	json.NewEncoder(w).Encode(response)
}

// restartInterfaceWindows disables and enables a network adapter using PowerShell
func restartInterfaceWindows(interfaceName string) error {
	disableCmd := exec.Command("powershell", "-Command", fmt.Sprintf("Disable-NetAdapter -Name \"%s\" -Confirm:$false", interfaceName))
	if out, err := disableCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to disable interface: %v, output: %s", err, string(out))
	}

	time.Sleep(1 * time.Second)

	enableCmd := exec.Command("powershell", "-Command", fmt.Sprintf("Enable-NetAdapter -Name \"%s\" -Confirm:$false", interfaceName))
	if out, err := enableCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to enable interface: %v, output: %s", err, string(out))
	}

	fmt.Printf("Successfully restarted interface %s\n", interfaceName)
	return nil
}

// sendError sends a JSON error response
