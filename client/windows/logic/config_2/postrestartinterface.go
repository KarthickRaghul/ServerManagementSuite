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

	interfaces, err := net.Interfaces()
	if err != nil {
		sendRestartError(w, "Failed to get network interfaces", err)
		return
	}

	var restartedInterfaces []string

	for _, iface := range interfaces {
		// Skip loopback and virtual interfaces
		if iface.Flags&net.FlagLoopback != 0 ||
			strings.Contains(iface.Name, "vEthernet") ||
			strings.Contains(iface.Name, "Loopback") ||
			strings.Contains(iface.Name, "Virtual") {
			continue
		}

		if iface.Flags&net.FlagUp != 0 {
			err := restartInterfaceWindows(iface.Name)
			if err != nil {
				fmt.Printf("Failed to restart interface %s: %v\n", iface.Name, err)
				continue
			}
			restartedInterfaces = append(restartedInterfaces, iface.Name)
		}
	}

	status := "failed"
	msg := "No interfaces were restarted"
	if len(restartedInterfaces) > 0 {
		status = "success"
		msg = fmt.Sprintf("Restarted %d interface(s)", len(restartedInterfaces))
	}

	resp := map[string]interface{}{
		"status":  status,
		"message": msg,
		"data": map[string]interface{}{
			"interfaces": restartedInterfaces,
		},
	}

	json.NewEncoder(w).Encode(resp)
}

// restartInterfaceWindows disables and re-enables a network adapter using PowerShell
func restartInterfaceWindows(interfaceName string) error {
	disableCmd := exec.Command("powershell", "-Command", fmt.Sprintf(`Disable-NetAdapter -Name "%s" -Confirm:$false`, interfaceName))
	if output, err := disableCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("disable error: %v, output: %s", err, output)
	}

	time.Sleep(1 * time.Second)

	enableCmd := exec.Command("powershell", "-Command", fmt.Sprintf(`Enable-NetAdapter -Name "%s" -Confirm:$false`, interfaceName))
	if output, err := enableCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("enable error: %v, output: %s", err, output)
	}

	fmt.Printf("Restarted interface: %s\n", interfaceName)
	return nil
}

// sendRestartError sends a standardized JSON error response
func sendRestartError(w http.ResponseWriter, msg string, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	resp := map[string]interface{}{
		"status":  "failed",
		"message": msg,
		"data": map[string]interface{}{
			"details": err.Error(),
		},
	}
	json.NewEncoder(w).Encode(resp)
}
