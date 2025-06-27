package config_2

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
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
	fmt.Println("Handling Windows network configuration update request...")

	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	var request NetworkUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		sendErrorResponse(w, "Failed to parse request body", err)
		return
	}

	// Validate request
	if request.Method != "static" && request.Method != "dynamic" {
		sendErrorResponse(w, "Invalid method, must be 'static' or 'dynamic'", nil)
		return
	}

	if request.Method == "static" {
		if request.IP == "" {
			sendErrorResponse(w, "IP address is required for static configuration", nil)
			return
		}
		// Validate IP format
		if net.ParseIP(request.IP) == nil {
			sendErrorResponse(w, "Invalid IP address format", nil)
			return
		}
	}

	// Get active interface
	iface, err := getActiveInterface()
	if err != nil {
		sendErrorResponse(w, "Failed to get active interface", err)
		return
	}

	log.Printf("Using interface: %s", iface)

	// Get current configuration
	oldConfig, err := getCurrentNetworkConfig(iface)
	if err != nil {
		log.Printf("Warning: Failed to get current config: %v", err)
		// Don't fail completely, create empty old config
		oldConfig = &NetworkUpdateRequest{
			Method:  "unknown",
			IP:      "",
			Subnet:  "",
			Gateway: "",
			DNS:     "",
		}
	}

	log.Printf("Old config: %+v", oldConfig)
	log.Printf("New config: %+v", request)

	var success bool
	var details string

	if request.Method == "dynamic" {
		success, details = setDynamicIP(iface)
	} else {
		// Fill in missing static configuration with defaults
		if request.Subnet == "" {
			if oldConfig.Subnet != "" {
				request.Subnet = oldConfig.Subnet
			} else {
				request.Subnet = "255.255.255.0" // Default subnet
			}
		}
		if request.Gateway == "" {
			if oldConfig.Gateway != "" {
				request.Gateway = oldConfig.Gateway
			} else {
				// Try to guess gateway from IP
				request.Gateway = guessGateway(request.IP, request.Subnet)
			}
		}
		if request.DNS == "" {
			if oldConfig.DNS != "" {
				request.DNS = oldConfig.DNS
			} else {
				request.DNS = "8.8.8.8,8.8.4.4" // Default DNS
			}
		}

		success, details = setStaticIP(iface, request.IP, request.Subnet, request.Gateway, request.DNS)
	}

	message := "Failed to update configuration"
	if success {
		message = "Configuration updated successfully"
	}

	resp := NetworkUpdateResponse{
		Success:   success,
		Message:   message,
		Details:   details,
		OldConfig: oldConfig,
		NewConfig: &request,
	}

	log.Printf("Response: %+v", resp)
	json.NewEncoder(w).Encode(resp)
}

func getActiveInterface() (string, error) {
	// Try multiple methods to get active interface

	// Method 1: Get interface with default route
	cmd := exec.Command("powershell", "-Command",
		`Get-NetRoute -DestinationPrefix "0.0.0.0/0" | Where-Object {$_.RouteMetric -ne $null} | Sort-Object RouteMetric | Select-Object -First 1 -ExpandProperty InterfaceAlias`)
	out, err := cmd.Output()
	if err == nil && strings.TrimSpace(string(out)) != "" {
		return strings.TrimSpace(string(out)), nil
	}

	// Method 2: Get interface with IP configuration
	cmd = exec.Command("powershell", "-Command",
		`Get-NetIPConfiguration | Where-Object {$_.IPv4DefaultGateway -ne $null} | Select-Object -First 1 -ExpandProperty InterfaceAlias`)
	out, err = cmd.Output()
	if err == nil && strings.TrimSpace(string(out)) != "" {
		return strings.TrimSpace(string(out)), nil
	}

	// Method 3: Get first non-loopback interface with IP
	cmd = exec.Command("powershell", "-Command",
		`Get-NetAdapter | Where-Object {$_.Status -eq "Up" -and $_.Name -notlike "*Loopback*"} | Select-Object -First 1 -ExpandProperty Name`)
	out, err = cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get active interface: %v", err)
	}

	iface := strings.TrimSpace(string(out))
	if iface == "" {
		return "", fmt.Errorf("no active interface found")
	}

	return iface, nil
}

func getCurrentNetworkConfig(iface string) (*NetworkUpdateRequest, error) {
	// Check if interface is using DHCP
	dhcpCmd := exec.Command("powershell", "-Command",
		fmt.Sprintf(`Get-NetIPInterface -InterfaceAlias "%s" -AddressFamily IPv4 | Select-Object -ExpandProperty Dhcp`, iface))
	dhcpOut, dhcpErr := dhcpCmd.Output()

	isDHCP := false
	if dhcpErr == nil {
		dhcpStatus := strings.TrimSpace(string(dhcpOut))
		isDHCP = strings.ToLower(dhcpStatus) == "enabled"
	}

	// Get IP configuration
	cmd := exec.Command("powershell", "-Command",
		fmt.Sprintf(`
		try {
			$ip = Get-NetIPAddress -InterfaceAlias "%s" -AddressFamily IPv4 -ErrorAction SilentlyContinue | Where-Object {$_.IPAddress -notlike "169.254.*"} | Select-Object -First 1
			$gw = Get-NetRoute -InterfaceAlias "%s" -DestinationPrefix "0.0.0.0/0" -ErrorAction SilentlyContinue | Select-Object -First 1
			$dns = (Get-DnsClientServerAddress -InterfaceAlias "%s" -AddressFamily IPv4 -ErrorAction SilentlyContinue).ServerAddresses
			
			$ipAddr = if ($ip) { $ip.IPAddress } else { "" }
			$prefix = if ($ip) { $ip.PrefixLength } else { "24" }
			$gateway = if ($gw) { $gw.NextHop } else { "" }
			$dnsStr = if ($dns) { $dns -join "," } else { "" }
			
			Write-Output "$ipAddr,$prefix,$gateway,$dnsStr"
		} catch {
			Write-Output ",,,"
		}`, iface, iface, iface))

	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get network config: %v", err)
	}

	parts := strings.Split(strings.TrimSpace(string(out)), ",")
	if len(parts) < 4 {
		return nil, fmt.Errorf("unexpected output format: %v", parts)
	}

	method := "static"
	if isDHCP {
		method = "dynamic"
	}

	config := &NetworkUpdateRequest{
		Method:  method,
		IP:      strings.TrimSpace(parts[0]),
		Subnet:  prefixToSubnet1(strings.TrimSpace(parts[1])),
		Gateway: strings.TrimSpace(parts[2]),
		DNS:     strings.TrimSpace(parts[3]),
	}

	// Handle empty values
	if config.IP == "" {
		config.IP = "0.0.0.0"
	}
	if config.Subnet == "" {
		config.Subnet = "255.255.255.0"
	}
	if config.Gateway == "" {
		config.Gateway = "0.0.0.0"
	}
	if config.DNS == "" {
		config.DNS = "8.8.8.8"
	}

	return config, nil
}

func setDynamicIP(iface string) (bool, string) {
	log.Printf("Setting DHCP for interface: %s", iface)

	// First, remove existing static IP configuration
	removeCmd := exec.Command("powershell", "-Command",
		fmt.Sprintf(`
		try {
			# Remove existing IP addresses (except DHCP ones)
			Get-NetIPAddress -InterfaceAlias "%s" -AddressFamily IPv4 -ErrorAction SilentlyContinue | Where-Object {$_.IPAddress -notlike "169.254.*"} | Remove-NetIPAddress -Confirm:$false -ErrorAction SilentlyContinue
			
			# Remove existing routes
			Get-NetRoute -InterfaceAlias "%s" -DestinationPrefix "0.0.0.0/0" -ErrorAction SilentlyContinue | Remove-NetRoute -Confirm:$false -ErrorAction SilentlyContinue
			
			Write-Output "Removed static configuration"
		} catch {
			Write-Output "Warning: $($_.Exception.Message)"
		}`, iface, iface))

	removeOut, _ := removeCmd.Output()
	log.Printf("Remove static config result: %s", string(removeOut))

	// Enable DHCP
	cmd := exec.Command("powershell", "-Command",
		fmt.Sprintf(`
		try {
			Set-NetIPInterface -InterfaceAlias "%s" -Dhcp Enabled -ErrorAction Stop
			Set-DnsClientServerAddress -InterfaceAlias "%s" -ResetServerAddresses -ErrorAction Stop
			
			# Restart network adapter to get new DHCP lease
			Restart-NetAdapter -Name "%s" -ErrorAction SilentlyContinue
			
			Write-Output "Successfully enabled DHCP"
		} catch {
			Write-Output "Error: $($_.Exception.Message)"
		}`, iface, iface, iface))

	out, err := cmd.Output()
	result := string(out)

	if err != nil {
		return false, fmt.Sprintf("Command failed: %v, Output: %s", err, result)
	}

	success := strings.Contains(result, "Successfully enabled DHCP")
	return success, result
}

func setStaticIP(iface, ip, subnet, gateway, dns string) (bool, string) {
	log.Printf("Setting static IP for interface: %s, IP: %s, Subnet: %s, Gateway: %s, DNS: %s",
		iface, ip, subnet, gateway, dns)

	prefix := subnetToPrefix(subnet)

	// First, disable DHCP and remove existing IP
	prepCmd := exec.Command("powershell", "-Command",
		fmt.Sprintf(`
		try {
			# Disable DHCP
			Set-NetIPInterface -InterfaceAlias "%s" -Dhcp Disabled -ErrorAction SilentlyContinue
			
			# Remove existing IP addresses
			Get-NetIPAddress -InterfaceAlias "%s" -AddressFamily IPv4 -ErrorAction SilentlyContinue | Remove-NetIPAddress -Confirm:$false -ErrorAction SilentlyContinue
			
			# Remove existing default routes
			Get-NetRoute -InterfaceAlias "%s" -DestinationPrefix "0.0.0.0/0" -ErrorAction SilentlyContinue | Remove-NetRoute -Confirm:$false -ErrorAction SilentlyContinue
			
			Write-Output "Prepared interface for static IP"
		} catch {
			Write-Output "Warning during preparation: $($_.Exception.Message)"
		}`, iface, iface, iface))

	prepOut, _ := prepCmd.Output()
	log.Printf("Preparation result: %s", string(prepOut))

	// Set static IP
	ipCmd := exec.Command("powershell", "-Command",
		fmt.Sprintf(`
		try {
			New-NetIPAddress -InterfaceAlias "%s" -IPAddress "%s" -PrefixLength %s -DefaultGateway "%s" -ErrorAction Stop
			Write-Output "SUCCESS: Static IP configured"
		} catch {
			Write-Output "ERROR: $($_.Exception.Message)"
		}`, iface, ip, prefix, gateway))

	var ipOut bytes.Buffer
	ipCmd.Stdout = &ipOut
	ipCmd.Stderr = &ipOut
	ipErr := ipCmd.Run()

	ipResult := ipOut.String()
	log.Printf("IP configuration result: %s", ipResult)

	if ipErr != nil || !strings.Contains(ipResult, "SUCCESS") {
		return false, fmt.Sprintf("Failed to set static IP: %s", ipResult)
	}

	// Set DNS
	if dns != "" {
		dnsCmd := exec.Command("powershell", "-Command",
			fmt.Sprintf(`
			try {
				Set-DnsClientServerAddress -InterfaceAlias "%s" -ServerAddresses %s -ErrorAction Stop
				Write-Output "SUCCESS: DNS configured"
			} catch {
				Write-Output "ERROR: $($_.Exception.Message)"
			}`, iface, formatDNSList(dns)))

		var dnsOut bytes.Buffer
		dnsCmd.Stdout = &dnsOut
		dnsCmd.Stderr = &dnsOut
		dnsErr := dnsCmd.Run()

		dnsResult := dnsOut.String()
		log.Printf("DNS configuration result: %s", dnsResult)

		if dnsErr != nil || !strings.Contains(dnsResult, "SUCCESS") {
			return true, fmt.Sprintf("Static IP set successfully, but DNS failed: %s", dnsResult)
		}
	}

	return true, "Successfully configured static IP and DNS"
}

func guessGateway(ip, subnet string) string {
	// Simple gateway guessing - typically .1 of the network
	ipParts := strings.Split(ip, ".")
	if len(ipParts) == 4 {
		return fmt.Sprintf("%s.%s.%s.1", ipParts[0], ipParts[1], ipParts[2])
	}
	return "192.168.1.1" // Fallback
}

func subnetToPrefix(subnet string) string {
	parts := strings.Split(subnet, ".")
	if len(parts) != 4 {
		return "24" // Default
	}

	bits := 0
	for _, part := range parts {
		n := 0
		fmt.Sscanf(part, "%d", &n)
		for i := 7; i >= 0; i-- {
			if n&(1<<i) != 0 {
				bits++
			} else {
				break
			}
		}
	}
	return fmt.Sprintf("%d", bits)
}

func prefixToSubnet1(prefix string) string {
	if prefix == "" {
		return "255.255.255.0" // Default
	}

	n := 0
	fmt.Sscanf(prefix, "%d", &n)
	if n < 0 || n > 32 {
		return "255.255.255.0" // Default for invalid prefix
	}

	mask := make([]int, 4)
	for i := 0; i < 4; i++ {
		if n >= 8 {
			mask[i] = 255
			n -= 8
		} else if n > 0 {
			mask[i] = ^(255 >> n) & 255
			n = 0
		} else {
			mask[i] = 0
		}
	}
	return fmt.Sprintf("%d.%d.%d.%d", mask[0], mask[1], mask[2], mask[3])
}

func formatDNSList(dns string) string {
	var quoted []string
	for _, s := range strings.Split(dns, ",") {
		trimmed := strings.TrimSpace(s)
		if trimmed != "" {
			quoted = append(quoted, fmt.Sprintf(`"%s"`, trimmed))
		}
	}
	if len(quoted) == 0 {
		return `"8.8.8.8"`
	}
	return strings.Join(quoted, ",")
}

func sendErrorResponse(w http.ResponseWriter, message string, err error) {
	w.WriteHeader(http.StatusBadRequest)
	resp := NetworkUpdateResponse{
		Success:   false,
		Message:   message,
		NewConfig: &NetworkUpdateRequest{}, // Ensure this is not nil
	}
	if err != nil {
		resp.Details = err.Error()
	}
	log.Printf("Error response: %s, Details: %s", message, resp.Details)
	json.NewEncoder(w).Encode(resp)
}
