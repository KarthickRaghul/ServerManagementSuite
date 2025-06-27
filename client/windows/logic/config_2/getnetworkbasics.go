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

// Standard response structures matching Linux implementation
type ErrorResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type SuccessResponse struct {
	Status string `json:"status"`
}

// Standard error response function - matches Linux exactly
func sendError(w http.ResponseWriter, message string, statusCode int) {
	w.WriteHeader(statusCode)
	errorResp := ErrorResponse{
		Status:  "failed",
		Message: message,
	}
	json.NewEncoder(w).Encode(errorResp)
}

// Standard success response function - matches Linux exactly
func sendPostSuccess(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
	response := SuccessResponse{
		Status: "success",
	}
	json.NewEncoder(w).Encode(response)
}

// HandleNetworkConfig handles Windows network config request
func HandleNetworkConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendError(w, "Only GET method allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	ip, iface, subnet, gateway := getIPAndGateway()
	dnsServers := getDNSServers()
	uptime := getSystemUptimeWindows()

	// ✅ Use the new reliable IP method detection
	ipMethod := getIPMethod(iface)

	response := NetworkConfigResponse{
		IPMethod:  ipMethod, // ✅ Now uses reliable detection
		IPAddress: ip,
		Gateway:   gateway,
		Subnet:    subnet,
		DNS:       strings.Join(dnsServers, ", "),
		Uptime:    uptime,
		Interface: make(map[string]InterfaceInfo),
	}

	// Get all network adapters
	allInterfaces := getAllNetworkAdaptersComplete()
	response.Interface = allInterfaces

	fmt.Println("Sending Windows network configuration response...")
	json.NewEncoder(w).Encode(response)
}

// HandleInterfaceControl handles enabling/disabling network interfaces
func HandleInterfaceControl(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, "Only POST method allowed", http.StatusMethodNotAllowed)
		return
	}

	var req InterfaceControlRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid JSON input: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Validate action
	if req.Action != "enable" && req.Action != "disable" {
		sendError(w, "Invalid action. Use 'enable' or 'disable'", http.StatusBadRequest)
		return
	}

	success := false
	message := ""

	switch req.Action {
	case "enable":
		success, message = enableNetworkInterface(req.InterfaceID)
	case "disable":
		success, message = disableNetworkInterface(req.InterfaceID)
	}

	if !success {
		sendError(w, message, http.StatusInternalServerError)
		return
	}

	// Send standard success response
	sendPostSuccess(w)
}

// ✅ NEW: Reliable IP method detection function
func getIPMethod(interfaceName string) string {
	if interfaceName == "" {
		return "static" // Default fallback
	}

	// Check DHCP status using Win32_NetworkAdapterConfiguration
	dhcpScript := fmt.Sprintf(`
	try {
		# Get the interface index
		$adapter = Get-NetAdapter -Name "%s" -ErrorAction Stop
		$interfaceIndex = $adapter.InterfaceIndex
		
		# Check DHCP configuration using Win32_NetworkAdapterConfiguration
		$dhcpConfig = Get-CimInstance -ClassName Win32_NetworkAdapterConfiguration | Where-Object { 
			$_.InterfaceIndex -eq $interfaceIndex
		}
		
		if ($dhcpConfig -and $dhcpConfig.DHCPEnabled -eq $true) {
			Write-Output "dynamic"
		} else {
			Write-Output "static"
		}
	} catch {
		# Fallback method: Check registry or netsh
		try {
			$netshOutput = netsh interface ip show config name="%s"
			if ($netshOutput -match "DHCP enabled:\s+Yes") {
				Write-Output "dynamic"
			} else {
				Write-Output "static"
			}
		} catch {
			Write-Output "static"
		}
	}
	`, interfaceName, interfaceName)

	cmd := exec.Command("powershell", "-Command", dhcpScript)
	output, err := cmd.Output()
	if err != nil {
		log.Printf("Error detecting IP method: %v", err)
		return "static" // Default fallback
	}

	result := strings.TrimSpace(string(output))
	if result == "dynamic" {
		return "dynamic"
	}
	return "static"
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
	if ($ipconfig) {
		$ip = $ipconfig.IPv4Address.IPAddress
		$iface = $ipconfig.InterfaceAlias
		$prefix = $ipconfig.IPv4Address.PrefixLength
		$gw = ($ipconfig.IPv4DefaultGateway.NextHop)
		Write-Output "$ip|$iface|$prefix|$gw"
	}
	`
	cmd := exec.Command("powershell", "-Command", script)
	output, err := cmd.Output()
	if err != nil {
		log.Printf("Error getting IP configuration: %v", err)
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

	if bits < 0 || bits > 32 {
		return "255.255.255.0" // Default subnet mask
	}

	mask := ^uint32(0) << (32 - bits)
	return fmt.Sprintf("%d.%d.%d.%d",
		(mask>>24)&0xFF,
		(mask>>16)&0xFF,
		(mask>>8)&0xFF,
		mask&0xFF)
}

// Get DNS servers
func getDNSServers() []string {
	script := `
	Get-DnsClientServerAddress -AddressFamily IPv4 | Where-Object { $_.ServerAddresses.Count -gt 0 } | ForEach-Object { 
		$_.ServerAddresses 
	} | Sort-Object | Get-Unique
	`
	cmd := exec.Command("powershell", "-Command", script)
	output, err := cmd.Output()
	if err != nil {
		log.Printf("Error getting DNS servers: %v", err)
		return []string{}
	}

	lines := strings.Split(string(output), "\n")
	servers := []string{}
	for _, line := range lines {
		ip := strings.TrimSpace(line)
		if net.ParseIP(ip) != nil && ip != "127.0.0.1" && ip != "::1" {
			servers = append(servers, ip)
		}
	}
	return servers
}

// Get system uptime
func getSystemUptimeWindows() string {
	script := `
	try {
		$boot = (Get-CimInstance Win32_OperatingSystem).LastBootUpTime
		$uptime = (Get-Date) - $boot
		"{0}d {1}h {2}m {3}s" -f $uptime.Days, $uptime.Hours, $uptime.Minutes, $uptime.Seconds
	} catch {
		"unknown"
	}
	`
	cmd := exec.Command("powershell", "-Command", script)
	output, err := cmd.Output()
	if err != nil {
		log.Printf("Error getting system uptime: %v", err)
		return "unknown"
	}
	return strings.TrimSpace(string(output))
}

// ✅ Optional: Force interface refresh after network changes
func refreshNetworkInterface(interfaceName string) {
	if interfaceName == "" {
		return
	}

	script := fmt.Sprintf(`
	try {
		# Clear DNS cache and refresh network configuration
		Clear-DnsClientCache
		
		# Restart the network adapter to refresh status
		Restart-NetAdapter -Name "%s" -Confirm:$false
		Start-Sleep -Seconds 2
		Write-Output "Interface refreshed successfully"
	} catch {
		Write-Output "Refresh failed: $($_.Exception.Message)"
	}
	`, interfaceName)

	cmd := exec.Command("powershell", "-Command", script)
	output, err := cmd.Output()
	if err != nil {
		log.Printf("Failed to refresh interface: %v", err)
	} else {
		log.Printf("Interface refresh result: %s", string(output))
	}
}
