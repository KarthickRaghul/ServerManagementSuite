package config_2

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"strings"
)

// InterfaceInfo represents information about a network interface
type InterfaceInfo struct {
	Mode   string `json:"mode"`
	Status string `json:"status"`
}

// NetworkConfigResponse represents the response format
type NetworkConfigResponse struct {
	IPMethod  string                   `json:"ip_method"`
	IPAddress string                   `json:"ip_address"`
	Gateway   string                   `json:"gateway"`
	Subnet    string                   `json:"subnet"`
	DNS       string                   `json:"dns"`
	Uptime    string                   `json:"uptime"`
	Interface map[string]InterfaceInfo `json:"interface"`
}

// HandleNetworkConfig handles the network config request on Windows
func HandleNetworkConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	response := NetworkConfigResponse{
		IPMethod:  "static",
		IPAddress: "",
		Gateway:   "",
		Subnet:    "",
		DNS:       "",
		Uptime:    "",
		Interface: make(map[string]InterfaceInfo),
	}

	ip, ifaceName, subnet, gateway := getIPAndGateway()
	response.IPAddress = ip
	response.Gateway = gateway
	response.Subnet = subnet
	if ip != "" {
		response.IPMethod = "dynamic"
	}

	ifaces, err := net.Interfaces()
	if err == nil {
		count := 1
		for _, iface := range ifaces {
			if (iface.Flags&net.FlagLoopback) != 0 || strings.Contains(iface.Name, "Loopback") {
				continue
			}
			status := "inactive"
			if iface.Name == ifaceName {
				status = "active"
			}
			response.Interface[fmt.Sprintf("%d", count)] = InterfaceInfo{
				Mode:   iface.Name,
				Status: status,
			}
			count++
		}
	}

	dnsServers := getDNSServers()
	response.DNS = strings.Join(dnsServers, ", ")

	response.Uptime = getSystemUptimeWindows()

	fmt.Println("Sending Windows network configuration response...")
	json.NewEncoder(w).Encode(response)
}

// getIPAndGateway returns the IP, interface name, subnet, and gateway
func getIPAndGateway() (string, string, string, string) {
	cmd := exec.Command("powershell", "(Get-NetIPConfiguration | Where-Object { $_.IPv4DefaultGateway -ne $null } | Select-Object -First 1).IPv4Address.IPAddress")
	out, err := cmd.Output()
	if err != nil {
		return "", "", "", ""
	}
	ip := strings.TrimSpace(string(out))

	cmd2 := exec.Command("powershell", "Get-NetIPConfiguration | Where-Object { $_.IPv4DefaultGateway -ne $null } | Select-Object -First 1 -ExpandProperty InterfaceAlias")
	ifaceOut, err := cmd2.Output()
	if err != nil {
		return ip, "", "", ""
	}
	iface := strings.TrimSpace(string(ifaceOut))

	cmd3 := exec.Command("powershell", "Get-NetIPConfiguration | Where-Object { $_.IPv4DefaultGateway -ne $null } | Select-Object -First 1 | ForEach-Object { $_.IPv4Address.PrefixLength }")
	subnetPrefix, err := cmd3.Output()
	subnet := "255.255.255.0"
	if err == nil {
		subnet = prefixToSubnet(strings.TrimSpace(string(subnetPrefix)))
	}

	cmd4 := exec.Command("powershell", "Get-NetRoute -DestinationPrefix 0.0.0.0/0 | Select-Object -First 1 -ExpandProperty NextHop")
	gatewayOut, err := cmd4.Output()
	gateway := strings.TrimSpace(string(gatewayOut))
	if err != nil {
		gateway = ""
	}

	return ip, iface, subnet, gateway
}

// Convert prefix length to subnet
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

// getDNSServers returns DNS servers from Windows
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

// getSystemUptimeWindows returns system uptime using PowerShell
func getSystemUptimeWindows() string {
	cmd := exec.Command("powershell", "-Command", `
		$uptime = (Get-CimInstance Win32_OperatingSystem).LastBootUpTime
		$uptimeSpan = (Get-Date) - $uptime
		"{0} days, {1} hours, {2} minutes" -f $uptimeSpan.Days, $uptimeSpan.Hours, $uptimeSpan.Minutes
	`)

	out, err := cmd.Output()
	if err != nil {
		return "Unable to determine uptime"
	}
	return strings.TrimSpace(string(out))
}
