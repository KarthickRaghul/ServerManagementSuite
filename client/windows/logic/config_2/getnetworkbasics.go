package config_2

import (
	"encoding/json"
	"fmt"
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

// HandleNetworkConfig handles Windows network config request
func HandleNetworkConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET method allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	ip, iface, subnet, gateway := getIPAndGateway()
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

	ifaces, err := net.Interfaces()
	if err == nil {
		index := 1
		for _, ifaceObj := range ifaces {
			if (ifaceObj.Flags&net.FlagLoopback) != 0 || strings.Contains(ifaceObj.Name, "Loopback") {
				continue
			}
			status := "inactive"
			if ifaceObj.Name == iface {
				status = "active"
			}
			power := "off"
			if ifaceObj.Flags&net.FlagUp != 0 {
				power = "on"
			}
			response.Interface[fmt.Sprintf("%d", index)] = InterfaceInfo{
				Mode:   ifaceObj.Name,
				Status: status,
				Power:  power,
			}
			index++
		}
	}

	fmt.Println("Sending Windows network configuration response...")
	json.NewEncoder(w).Encode(response)
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
