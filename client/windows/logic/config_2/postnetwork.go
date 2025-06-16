package config_2

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"strings"
)

type NetworkUpdateRequest struct {
	Method  string `json:"method"`            // "static" or "dynamic"
	IP      string `json:"ip,omitempty"`      // Required for static
	Subnet  string `json:"subnet,omitempty"`  // Optional fallback to current
	Gateway string `json:"gateway,omitempty"` // Optional fallback to current
	DNS     string `json:"dns,omitempty"`     // Comma-separated DNS
}

func HandleUpdateNetworkConfig(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Handling network configuration update request on Windows...")

	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	var request NetworkUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		sendErrorResponse(w, "Failed to parse request body", err, nil)
		return
	}

	if request.Method != "static" && request.Method != "dynamic" {
		sendErrorResponse(w, "Invalid method, must be 'static' or 'dynamic'", nil, &request)
		return
	}

	if request.Method == "static" && request.IP == "" {
		sendErrorResponse(w, "IP address is required for static configuration", nil, &request)
		return
	}

	iface, err := getActiveInterface()
	if err != nil {
		sendErrorResponse(w, "Failed to determine active interface", err, &request)
		return
	}

	oldConfig, err := getCurrentNetworkConfigWindows(iface)
	if err != nil {
		sendErrorResponse(w, "Failed to get current configuration", err, &request)
		return
	}

	// Fallback: Fill missing values from old config
	if request.Method == "static" {
		if request.Subnet == "" {
			request.Subnet = oldConfig.Subnet
		}
		if request.Gateway == "" {
			request.Gateway = oldConfig.Gateway
		}
		if request.DNS == "" {
			request.DNS = oldConfig.DNS
		}
	}

	var success bool
	var details string
	if request.Method == "dynamic" {
		success, details = setDynamicIPWindows(iface)
	} else {
		success, details = setStaticIPWindows(iface, request.IP, request.Subnet, request.Gateway, request.DNS)
	}

	status := "failed"
	msg := "Failed to update network configuration"
	if success {
		status = "success"
		msg = "Network configuration updated successfully"
	}

	resp := map[string]interface{}{
		"status":  status,
		"message": msg,
		"data": map[string]interface{}{
			"details":    details,
			"old_config": oldConfig,
			"new_config": request,
		},
	}
	json.NewEncoder(w).Encode(resp)
}

// Powershell command to get the active interface alias (lowest metric default route)
func getActiveInterface() (string, error) {
	cmd := exec.Command("powershell", "-Command", "Get-NetRoute -DestinationPrefix 0.0.0.0/0 | Sort-Object RouteMetric | Select-Object -First 1 -ExpandProperty InterfaceAlias")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func getCurrentNetworkConfigWindows(iface string) (*NetworkUpdateRequest, error) {
	// IP
	cmd := exec.Command("powershell", "-Command", fmt.Sprintf(`(Get-NetIPAddress -InterfaceAlias "%s" -AddressFamily IPv4 | Where-Object {$_.IPAddress -ne $null} | Select-Object -First 1).IPAddress`, iface))
	ip, _ := cmd.Output()

	// Prefix
	cmd = exec.Command("powershell", "-Command", fmt.Sprintf(`(Get-NetIPAddress -InterfaceAlias "%s" -AddressFamily IPv4 | Where-Object {$_.PrefixLength -ne $null} | Select-Object -First 1).PrefixLength`, iface))
	prefix, _ := cmd.Output()

	// Gateway
	cmd = exec.Command("powershell", "-Command", fmt.Sprintf(`(Get-NetRoute -InterfaceAlias "%s" -DestinationPrefix 0.0.0.0/0 | Select-Object -First 1).NextHop`, iface))
	gw, _ := cmd.Output()

	// DNS
	cmd = exec.Command("powershell", "-Command", fmt.Sprintf(`(Get-DnsClientServerAddress -InterfaceAlias "%s" -AddressFamily IPv4).ServerAddresses -join ","`, iface))
	dns, _ := cmd.Output()

	return &NetworkUpdateRequest{
		Method:  "static", // DHCP not detected here
		IP:      strings.TrimSpace(string(ip)),
		Subnet:  prefixToNetmask(strings.TrimSpace(string(prefix))),
		Gateway: strings.TrimSpace(string(gw)),
		DNS:     strings.TrimSpace(string(dns)),
	}, nil
}

func setDynamicIPWindows(iface string) (bool, string) {
	cmd1 := fmt.Sprintf(`netsh interface ip set address name="%s" source=dhcp`, iface)
	cmd2 := fmt.Sprintf(`netsh interface ip set dns name="%s" source=dhcp`, iface)

	if err := exec.Command("cmd", "/C", cmd1).Run(); err != nil {
		return false, "Failed to switch to DHCP (IP)"
	}
	if err := exec.Command("cmd", "/C", cmd2).Run(); err != nil {
		return false, "IP set to DHCP, but DNS change failed"
	}
	return true, "Set to dynamic IP using netsh"
}

func setStaticIPWindows(iface, ip, subnet, gateway, dns string) (bool, string) {
	cmd := fmt.Sprintf(`netsh interface ip set address name="%s" static %s %s %s`, iface, ip, subnet, gateway)
	if err := exec.Command("cmd", "/C", cmd).Run(); err != nil {
		return false, "Failed to set static IP"
	}

	// Set DNS (may be multiple)
	if dns != "" {
		for i, d := range strings.Split(dns, ",") {
			dnsCmd := fmt.Sprintf(`netsh interface ip add dns name="%s" %s %s`, iface, d, func() string {
				if i == 0 {
					return "validate=no"
				}
				return "index=2"
			}())
			if err := exec.Command("cmd", "/C", dnsCmd).Run(); err != nil {
				return false, "IP set, DNS failed: " + err.Error()
			}
		}
	}
	return true, "Successfully applied static IP"
}

// Convert /24 to 255.255.255.0
func prefixToNetmask(prefix string) string {
	var p int
	fmt.Sscanf(prefix, "%d", &p)
	mask := net.CIDRMask(p, 32)
	return fmt.Sprintf("%d.%d.%d.%d", mask[0], mask[1], mask[2], mask[3])
}

func sendErrorResponse(w http.ResponseWriter, msg string, err error, req *NetworkUpdateRequest) {
	w.WriteHeader(http.StatusBadRequest)

	data := map[string]interface{}{
		"new_config": req,
	}
	if err != nil {
		data["details"] = err.Error()
	}

	resp := map[string]interface{}{
		"status":  "failed",
		"message": msg,
		"data":    data,
	}

	json.NewEncoder(w).Encode(resp)
}
