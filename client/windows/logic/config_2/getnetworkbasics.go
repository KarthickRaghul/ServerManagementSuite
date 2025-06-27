package config_2

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os/exec"
	"strings"
)

// InterfaceInfo represents network interface details
type InterfaceInfo struct {
	Mode   string `json:"mode"`
	Status string `json:"status"`
	Power  string `json:"power"`
}

// NetworkConfigResponse mirrors Linux structure
type NetworkConfigResponse struct {
	IPMethod  string                   `json:"ip_method"`
	IPAddress string                   `json:"ip_address"`
	Gateway   string                   `json:"gateway"`
	Subnet    string                   `json:"subnet"`
	DNS       string                   `json:"dns"`
	Uptime    string                   `json:"uptime"`
	Interface map[string]InterfaceInfo `json:"interface"`
}

// InterfaceControlRequest for enabling/disabling interfaces
type InterfaceControlRequest struct {
	InterfaceID string `json:"interface_id"`
	Action      string `json:"action"` // "enable" or "disable"
}

// HandleNetworkConfig handles Windows network config request
func HandleNetworkConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET method allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	ip, _, subnet, gateway := getIPAndGateway()
	dnsServers := getDNSServers()
	uptime := getSystemUptimeWindows()

	response := NetworkConfigResponse{
		IPMethod:  "static",
		IPAddress: ip,
		Gateway:   gateway,
		Subnet:    subnet,
		DNS:       strings.Join(dnsServers, ", "),
		Uptime:    uptime,
		Interface: make(map[string]InterfaceInfo),
	}

	if ip != "" {
		response.IPMethod = "dynamic"
	}

	// Get all network adapters (including disabled ones) using PowerShell
	allInterfaces := getAllNetworkAdaptersComplete()

	response.Interface = allInterfaces

	fmt.Println("Sending Windows network configuration response...")
	json.NewEncoder(w).Encode(response)
}

// HandleInterfaceControl handles enabling/disabling network interfaces
func HandleInterfaceControl(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method allowed", http.StatusMethodNotAllowed)
		return
	}

	var req InterfaceControlRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	success := false
	message := ""

	switch req.Action {
	case "enable":
		success, message = enableNetworkInterface(req.InterfaceID)
	case "disable":
		success, message = disableNetworkInterface(req.InterfaceID)
	default:
		http.Error(w, "Invalid action. Use 'enable' or 'disable'", http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"success": success,
		"message": message,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Get all network adapters including disabled ones - maintains same format
func getAllNetworkAdaptersComplete() map[string]InterfaceInfo {
	script := `
	Get-NetAdapter | ForEach-Object {
		$status = if ($_.Status -eq "Up") { "active" } else { "inactive" }
		$power = if ($_.AdminStatus -eq "Up") { "on" } else { "off" }
		Write-Output "$($_.Name)|$status|$power"
	}
	`

	cmd := exec.Command("powershell", "-Command", script)
	output, err := cmd.Output()
	if err != nil {
		log.Printf("Error getting network adapters: %v", err)
		return make(map[string]InterfaceInfo)
	}

	interfaces := make(map[string]InterfaceInfo)
	lines := strings.Split(string(output), "\n")
	index := 1

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Split(line, "|")
		if len(parts) >= 3 {
			// Skip loopback and virtual adapters
			interfaceName := strings.TrimSpace(parts[0])
			if strings.Contains(strings.ToLower(interfaceName), "loopback") {
				continue
			}

			interfaceInfo := InterfaceInfo{
				Mode:   interfaceName,
				Status: strings.TrimSpace(parts[1]),
				Power:  strings.TrimSpace(parts[2]),
			}
			interfaces[fmt.Sprintf("%d", index)] = interfaceInfo
			index++
		}
	}

	return interfaces
}

// Enable network interface by name or device ID
func enableNetworkInterface(interfaceID string) (bool, string) {
	// Try by interface name first
	script := fmt.Sprintf(`
	try {
		$adapter = Get-NetAdapter | Where-Object { $_.Name -eq "%s" -or $_.DeviceID -eq "%s" } | Select-Object -First 1
		if ($adapter) {
			Enable-NetAdapter -Name $adapter.Name -Confirm:$false
			"SUCCESS: Interface enabled"
		} else {
			"ERROR: Interface not found"
		}
	} catch {
		"ERROR: $($_.Exception.Message)"
	}
	`, interfaceID, interfaceID)

	cmd := exec.Command("powershell", "-Command", script)
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Sprintf("Command execution failed: %v", err)
	}

	result := strings.TrimSpace(string(output))
	success := strings.HasPrefix(result, "SUCCESS")

	if success {
		log.Printf("Successfully enabled interface: %s", interfaceID)
	} else {
		log.Printf("Failed to enable interface %s: %s", interfaceID, result)
	}

	return success, result
}

// Disable network interface by name or device ID
func disableNetworkInterface(interfaceID string) (bool, string) {
	script := fmt.Sprintf(`
	try {
		$adapter = Get-NetAdapter | Where-Object { $_.Name -eq "%s" -or $_.DeviceID -eq "%s" } | Select-Object -First 1
		if ($adapter) {
			Disable-NetAdapter -Name $adapter.Name -Confirm:$false
			"SUCCESS: Interface disabled"
		} else {
			"ERROR: Interface not found"
		}
	} catch {
		"ERROR: $($_.Exception.Message)"
	}
	`, interfaceID, interfaceID)

	cmd := exec.Command("powershell", "-Command", script)
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Sprintf("Command execution failed: %v", err)
	}

	result := strings.TrimSpace(string(output))
	success := strings.HasPrefix(result, "SUCCESS")

	if success {
		log.Printf("Successfully disabled interface: %s", interfaceID)
	} else {
		log.Printf("Failed to disable interface %s: %s", interfaceID, result)
	}

	return success, result
}

// Get IP, subnet, gateway, interface name
func getIPAndGateway() (ip, iface, subnet, gateway string) {
	script := `
	$ipconfig = Get-NetIPConfiguration | Where-Object { $_.IPv4DefaultGateway -ne $null } | Select-Object -First 1
	$ip = $ipconfig.IPv4Address.IPAddress
	$iface = $ipconfig.InterfaceAlias
	$prefix = $ipconfig.IPv4Address.PrefixLength
	$gw = ($ipconfig.IPv4DefaultGateway.NextHop)
	Write-Output "$ip|$iface|$prefix|$gw"
	`
	cmd := exec.Command("powershell", "-Command", script)
	output, err := cmd.Output()
	if err != nil {
		return
	}

	parts := strings.Split(strings.TrimSpace(string(output)), "|")
	if len(parts) == 4 {
		ip = strings.TrimSpace(parts[0])
		iface = strings.TrimSpace(parts[1])
		subnet = prefixToSubnet(parts[2])
		gateway = strings.TrimSpace(parts[3])
	}
	return
}

// Convert CIDR prefix to subnet mask
func prefixToSubnet(prefix string) string {
	var bits int
	fmt.Sscanf(prefix, "%d", &bits)
	mask := ^uint32(0) << (32 - bits)
	return fmt.Sprintf("%d.%d.%d.%d",
		(mask>>24)&0xFF,
		(mask>>16)&0xFF,
		(mask>>8)&0xFF,
		mask&0xFF)
}

// Get DNS servers
func getDNSServers() []string {
	cmd := exec.Command("powershell", "Get-DnsClientServerAddress -AddressFamily IPv4 | ForEach-Object { $_.ServerAddresses }")
	output, err := cmd.Output()
	if err != nil {
		return []string{}
	}
	lines := strings.Split(string(output), "\n")
	servers := []string{}
	for _, line := range lines {
		ip := strings.TrimSpace(line)
		if net.ParseIP(ip) != nil {
			servers = append(servers, ip)
		}
	}
	return servers
}

// Get system uptime
func getSystemUptimeWindows() string {
	script := `
	$boot = (Get-CimInstance Win32_OperatingSystem).LastBootUpTime
	$uptime = (Get-Date) - $boot
	"{0}d {1}h {2}m {3}s" -f $uptime.Days, $uptime.Hours, $uptime.Minutes, $uptime.Seconds
	`
	cmd := exec.Command("powershell", "-Command", script)
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(output))
}

