package config_2

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

type NetworkUpdateRequest struct {
	Method  string `json:"method"`
	IP      string `json:"ip,omitempty"`
	Subnet  string `json:"subnet,omitempty"`
	Gateway string `json:"gateway,omitempty"`
	DNS     string `json:"dns,omitempty"`
}

type NetworkUpdateResponse struct {
	Success   bool                  `json:"success"`
	Message   string                `json:"message"`
	Details   string                `json:"details,omitempty"`
	OldConfig *NetworkUpdateRequest `json:"old_config,omitempty"`
	NewConfig *NetworkUpdateRequest `json:"new_config"`
}

func HandleUpdateNetworkConfig(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Current Date and Time (UTC):", time.Now().UTC().Format("2006-01-02 15:04:05"))
	fmt.Println("Handling network configuration update request...")

	// Check for POST method
	if r.Method != http.MethodPost {
		sendError(w, "Only POST method allowed", http.StatusMethodNotAllowed)
		return
	}

	// Set content type
	w.Header().Set("Content-Type", "application/json")

	// Parse request body
	var request NetworkUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		sendError(w, "Failed to parse request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Validate method
	if request.Method != "static" && request.Method != "dynamic" {
		sendError(w, "Invalid method, must be 'static' or 'dynamic'", http.StatusBadRequest)
		return
	}

	// Validate static configuration requirements
	if request.Method == "static" && request.IP == "" {
		sendError(w, "IP address is required for static configuration", http.StatusBadRequest)
		return
	}

	// Get active interface
	activeInterface, err := getActiveInterface()
	if err != nil {
		sendError(w, "Failed to get active network interface: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Get current configuration
	oldConfig, err := getCurrentNetworkConfig(activeInterface)
	if err != nil {
		sendError(w, "Failed to get current network configuration: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Apply network configuration
	var details string
	var success bool
	if request.Method == "dynamic" {
		success, details = setDynamicIP(activeInterface)
	} else {
		// Fill in missing static configuration with current values
		if request.Subnet == "" {
			request.Subnet = oldConfig.Subnet
		}
		if request.Gateway == "" {
			request.Gateway = oldConfig.Gateway
		}
		if request.DNS == "" {
			request.DNS = oldConfig.DNS
		}
		success, details = setStaticIP(activeInterface, request.IP, request.Subnet, request.Gateway, request.DNS)
	}

	// Check if configuration was successful
	if !success {
		sendError(w, "Failed to update network configuration: "+details, http.StatusInternalServerError)
		return
	}

	// Send success response
	sendPostSuccess(w)
}


func getCurrentNetworkConfig(iface string) (*NetworkUpdateRequest, error) {
	config := &NetworkUpdateRequest{}

	isUsingDHCP, err := isInterfaceUsingDHCP(iface)
	if err != nil {
		return nil, err
	}
	if isUsingDHCP {
		config.Method = "dynamic"
	} else {
		config.Method = "static"
	}

	ipOutput, err := exec.Command("ip", "-o", "-4", "addr", "show", iface).Output()
	if err == nil {
		ipLine := string(ipOutput)
		ipMatch := strings.Fields(ipLine)
		if len(ipMatch) > 3 {
			ipCIDR := ipMatch[3]
			ipParts := strings.Split(ipCIDR, "/")
			if len(ipParts) > 0 {
				config.IP = ipParts[0]
				if len(ipParts) > 1 {
					config.Subnet = cidrToNetmask(ipParts[1])
				}
			}
		}
	}

	gatewayOutput, err := exec.Command("ip", "route", "show", "dev", iface).Output()
	if err == nil {
		for _, line := range strings.Split(string(gatewayOutput), "\n") {
			if strings.Contains(line, "default via") {
				fields := strings.Fields(line)
				if len(fields) > 2 {
					config.Gateway = fields[2]
					break
				}
			}
		}
	}

	dnsServers, err := getDNSServers()
	if err == nil && len(dnsServers) > 0 {
		config.DNS = strings.Join(dnsServers, ",")
	}

	return config, nil
}

func isInterfaceUsingDHCP(iface string) (bool, error) {
	psOutput, err := exec.Command("ps", "-ef").Output()
	if err != nil {
		return false, err
	}
	psStr := string(psOutput)
	if strings.Contains(psStr, "dhclient "+iface) || strings.Contains(psStr, "dhcpcd "+iface) {
		return true, nil
	}
	if _, err := exec.LookPath("nmcli"); err == nil {
		nmOutput, err := exec.Command("nmcli", "-g", "IP4.METHOD", "device", "show", iface).Output()
		if err == nil && strings.Contains(string(nmOutput), "auto") {
			return true, nil
		}
	}
	return false, nil
}

func getActiveInterface() (string, error) {
	output, err := exec.Command("ip", "route", "show", "default").Output()
	if err != nil {
		return "", fmt.Errorf("failed to get default route: %w", err)
	}
	parts := strings.Fields(string(output))
	for i, part := range parts {
		if part == "dev" && i+1 < len(parts) {
			return parts[i+1], nil
		}
	}
	return "", fmt.Errorf("could not determine active interface")
}

func getConnectionNameForInterface(iface string) (string, error) {
	output, err := exec.Command("nmcli", "-t", "-f", "DEVICE,NAME", "connection", "show").Output()
	if err != nil {
		return "", err
	}
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		fields := strings.Split(line, ":")
		if len(fields) == 2 && fields[0] == iface {
			return fields[1], nil
		}
	}
	return "", fmt.Errorf("no connection name found for interface %s", iface)
}

func setDynamicIP(iface string) (bool, string) {
	if _, err := exec.LookPath("nmcli"); err == nil {
		conName, err := getConnectionNameForInterface(iface)
		if err == nil {
			cmd := exec.Command("nmcli", "con", "mod", conName, "ipv4.method", "auto", "ipv4.addresses", "", "ipv4.gateway", "", "ipv4.dns", "")
			if out, err := cmd.CombinedOutput(); err != nil {
				return false, fmt.Sprintf("Failed to set dynamic IP with nmcli: %s: %s", err, out)
			}
			upCmd := exec.Command("nmcli", "con", "up", conName)
			if upOut, err := upCmd.CombinedOutput(); err != nil {
				return false, fmt.Sprintf("Failed to bring up connection: %s: %s", err, upOut)
			}
			return true, "Successfully set dynamic IP using NetworkManager"
		}
	}

	_ = exec.Command("dhclient", "-r", iface).Run()
	cmd := exec.Command("dhclient", iface)
	if out, err := cmd.CombinedOutput(); err != nil {
		return false, fmt.Sprintf("Failed to run dhclient: %s: %s", err, out)
	}
	return true, "Successfully set dynamic IP using dhclient"
}

func setStaticIP(iface, ip, subnet, gateway, dns string) (bool, string) {
	if _, err := exec.LookPath("nmcli"); err == nil {
		conName, err := getConnectionNameForInterface(iface)
		if err == nil {
			args := []string{"con", "mod", conName, "ipv4.method", "manual", "ipv4.addresses", fmt.Sprintf("%s/%s", ip, subnetToPrefix(subnet))}
			if gateway != "" {
				args = append(args, "ipv4.gateway", gateway)
			}
			if dns != "" {
				args = append(args, "ipv4.dns", dns)
			}
			cmd := exec.Command("nmcli", args...)
			if out, err := cmd.CombinedOutput(); err != nil {
				return false, fmt.Sprintf("Failed to set static IP using nmcli: %s: %s", err, out)
			}
			upCmd := exec.Command("nmcli", "con", "up", conName)
			if upOut, err := upCmd.CombinedOutput(); err != nil {
				return false, fmt.Sprintf("Failed to apply static IP: %s: %s", err, upOut)
			}
			return true, "Successfully set static IP using NetworkManager"
		}
	}

	_ = exec.Command("ip", "addr", "flush", "dev", iface).Run()
	addCmd := exec.Command("ip", "addr", "add", fmt.Sprintf("%s/%s", ip, subnetToPrefix(subnet)), "dev", iface)
	if out, err := addCmd.CombinedOutput(); err != nil {
		return false, fmt.Sprintf("Failed to set static IP: %s: %s", err, out)
	}
	if gateway != "" {
		_ = exec.Command("ip", "route", "del", "default").Run()
		addRouteCmd := exec.Command("ip", "route", "add", "default", "via", gateway, "dev", iface)
		if out, err := addRouteCmd.CombinedOutput(); err != nil {
			return false, fmt.Sprintf("Failed to set gateway: %s: %s", err, out)
		}
	}
	if dns != "" {
		content := ""
		for _, s := range strings.Split(dns, ",") {
			content += fmt.Sprintf("nameserver %s\n", strings.TrimSpace(s))
		}
		if err := os.WriteFile("/etc/resolv.conf", []byte(content), 0644); err != nil {
			return true, "IP and gateway set. Failed to update DNS: " + err.Error()
		}
	}
	return true, "Successfully set static IP configuration"
}

func subnetToPrefix(subnet string) string {
	octets := strings.Split(subnet, ".")
	count := 0
	for _, oct := range octets {
		n := 0
		fmt.Sscanf(oct, "%d", &n)
		for i := 7; i >= 0; i-- {
			if n&(1<<uint(i)) != 0 {
				count++
			} else {
				break
			}
		}
	}
	return fmt.Sprintf("%d", count)
}

func sendErrorResponse(w http.ResponseWriter, message string, err error) {
	w.WriteHeader(http.StatusBadRequest)
	errResp := NetworkUpdateResponse{
		Success: false,
		Message: message,
		Details: "",
	}
	if err != nil {
		errResp.Details = err.Error()
	}
	json.NewEncoder(w).Encode(errResp)
}
