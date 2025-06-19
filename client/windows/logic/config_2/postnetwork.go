package config_2

import (
	"bytes"
	"encoding/json"
	"fmt"
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

	if request.Method != "static" && request.Method != "dynamic" {
		sendErrorResponse(w, "Invalid method, must be 'static' or 'dynamic'", nil)
		return
	}

	if request.Method == "static" && request.IP == "" {
		sendErrorResponse(w, "IP address is required for static configuration", nil)
		return
	}

	iface, err := getActiveInterface()
	if err != nil {
		sendErrorResponse(w, "Failed to get active interface", err)
		return
	}

	oldConfig, err := getCurrentNetworkConfig(iface)
	if err != nil {
		sendErrorResponse(w, "Failed to fetch current config", err)
		return
	}

	var success bool
	var details string

	if request.Method == "dynamic" {
		success, details = setDynamicIP(iface)
	} else {
		if request.Subnet == "" {
			request.Subnet = oldConfig.Subnet
		}
		if request.Gateway == "" {
			request.Gateway = oldConfig.Gateway
		}
		if request.DNS == "" {
			request.DNS = oldConfig.DNS
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
	json.NewEncoder(w).Encode(resp)
}

func getActiveInterface() (string, error) {
	cmd := exec.Command("powershell", "-Command", `Get-NetRoute -DestinationPrefix "0.0.0.0/0" | Sort-Object RouteMetric | Select-Object -First 1 -ExpandProperty InterfaceAlias`)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func getCurrentNetworkConfig(iface string) (*NetworkUpdateRequest, error) {
	cmd := exec.Command("powershell", "-Command",
		fmt.Sprintf(`$ip = Get-NetIPAddress -InterfaceAlias "%s" -AddressFamily IPv4 | Select-Object -First 1; 
			$gw = Get-NetRoute -InterfaceAlias "%s" -DestinationPrefix "0.0.0.0/0" | Select-Object -First 1;
			$dns = (Get-DnsClientServerAddress -InterfaceAlias "%s" -AddressFamily IPv4).ServerAddresses -join ",";
			Write-Output "$($ip.IPAddress),$($ip.PrefixLength),$($gw.NextHop),$dns"`,
			iface, iface, iface),
	)

	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	parts := strings.Split(strings.TrimSpace(string(out)), ",")
	if len(parts) < 4 {
		return nil, fmt.Errorf("unexpected output format: %v", parts)
	}

	return &NetworkUpdateRequest{
		Method:  "static", // Windows doesn't expose DHCP state easily via this method
		IP:      parts[0],
		Subnet:  prefixToSubnet1(parts[1]),
		Gateway: parts[2],
		DNS:     parts[3],
	}, nil
}

func setDynamicIP(iface string) (bool, string) {
	cmd := exec.Command("powershell", "-Command",
		fmt.Sprintf(`Set-NetIPInterface -InterfaceAlias "%s" -Dhcp Enabled;
			Set-DnsClientServerAddress -InterfaceAlias "%s" -ResetServerAddresses`, iface, iface))
	out, err := cmd.CombinedOutput()
	if err != nil {
		return false, string(out)
	}
	return true, string(out)
}

func setStaticIP(iface, ip, subnet, gateway, dns string) (bool, string) {
	prefix := subnetToPrefix(subnet)

	cmd := exec.Command("powershell", "-Command",
		fmt.Sprintf(`New-NetIPAddress -InterfaceAlias "%s" -IPAddress "%s" -PrefixLength %s -DefaultGateway "%s" -ErrorAction Stop`,
			iface, ip, prefix, gateway))
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	if err != nil {
		return false, out.String()
	}

	dnsCmd := exec.Command("powershell", "-Command",
		fmt.Sprintf(`Set-DnsClientServerAddress -InterfaceAlias "%s" -ServerAddresses (%s)`,
			iface, formatDNSList(dns)))
	var dnsOut bytes.Buffer
	dnsCmd.Stdout = &dnsOut
	dnsCmd.Stderr = &dnsOut
	err = dnsCmd.Run()
	if err != nil {
		return true, "Static IP set, but DNS failed: " + dnsOut.String()
	}

	return true, "Successfully set static configuration"
}

func subnetToPrefix(subnet string) string {
	parts := strings.Split(subnet, ".")
	bits := 0
	for _, part := range parts {
		n := 0
		fmt.Sscanf(part, "%d", &n)
		for i := 7; i >= 0; i-- {
			if n&(1<<i) != 0 {
				bits++
			}
		}
	}
	return fmt.Sprintf("%d", bits)
}

func prefixToSubnet1(prefix string) string {
	n := 0
	fmt.Sscanf(prefix, "%d", &n)
	mask := make([]int, 4)
	for i := 0; i < 4; i++ {
		if n >= 8 {
			mask[i] = 255
			n -= 8
		} else {
			mask[i] = ^(255 >> n) & 255
			n = 0
		}
	}
	return fmt.Sprintf("%d.%d.%d.%d", mask[0], mask[1], mask[2], mask[3])
}

func formatDNSList(dns string) string {
	var quoted []string
	for _, s := range strings.Split(dns, ",") {
		quoted = append(quoted, fmt.Sprintf(`"%s"`, strings.TrimSpace(s)))
	}
	return strings.Join(quoted, ",")
}

func sendErrorResponse(w http.ResponseWriter, message string, err error) {
	w.WriteHeader(http.StatusBadRequest)
	resp := NetworkUpdateResponse{
		Success: false,
		Message: message,
	}
	if err != nil {
		resp.Details = err.Error()
	}
	json.NewEncoder(w).Encode(resp)
}
