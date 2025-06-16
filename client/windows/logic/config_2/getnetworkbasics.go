package config_2

import (
	"bufio"
	"encoding/json"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
)

type InterfaceStatus struct {
	Mode   string `json:"mode"`
	Status string `json:"status"`
}

type NetworkConfig struct {
	IPMethod  string                     `json:"ip_method"`
	IPAddress string                     `json:"ip_address"`
	Gateway   string                     `json:"gateway"`
	Subnet    string                     `json:"subnet"`
	DNS       string                     `json:"dns"`
	Interface map[string]InterfaceStatus `json:"interface"`
}

type NetAdapter struct {
	Name   string `json:"Name"`
	Status string `json:"Status"`
}

func NetworkConfigHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "failed",
			"message": "Only GET method allowed",
		})
		return
	}

	config := NetworkConfig{
		Interface: make(map[string]InterfaceStatus),
	}

	cmd := exec.Command("netsh", "interface", "ip", "show", "config")
	output, err := cmd.Output()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "failed",
			"message": "Failed to get IP config: " + err.Error(),
		})
		return
	}

	scanner := bufio.NewScanner(strings.NewReader(string(output)))

	var (
		tempMethod string
		tempIP     string
		tempGW     string
		tempSubnet string
		tempDNS    []string
		captured   bool
		isLoopback bool
	)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, "Configuration for interface") {
			if !captured && tempIP != "" && !isLoopback {
				config.IPMethod = tempMethod
				config.IPAddress = tempIP
				config.Gateway = tempGW
				config.Subnet = tempSubnet
				config.DNS = strings.Join(tempDNS, ", ")
				captured = true
			}
			tempMethod, tempIP, tempGW, tempSubnet = "", "", "", ""
			tempDNS = []string{}
			isLoopback = false
		} else if strings.HasPrefix(line, "DHCP enabled") {
			if strings.Contains(line, "Yes") {
				tempMethod = "dynamic"
			} else {
				tempMethod = "static"
			}
		} else if strings.HasPrefix(line, "IP Address") || strings.HasPrefix(line, "IPv4 Address") {
			ip := strings.TrimSpace(strings.Split(line, ":")[1])
			ip = strings.Split(ip, "(")[0]
			ip = strings.TrimSpace(ip)
			if ip == "127.0.0.1" {
				isLoopback = true
			}
			tempIP = ip
		} else if strings.HasPrefix(line, "Default Gateway") && strings.Contains(line, ":") {
			tempGW = strings.TrimSpace(strings.Split(line, ":")[1])
		} else if strings.HasPrefix(line, "Subnet Prefix") && strings.Contains(line, "(") {
			subnet := strings.TrimSpace(line[strings.LastIndex(line, "(")+1:])
			subnet = strings.TrimSuffix(subnet, ")")
			tempSubnet = subnet
		} else if strings.HasPrefix(line, "DNS servers configured through DHCP") ||
			strings.HasPrefix(line, "Statically Configured DNS Servers") {

			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				dns := strings.TrimSpace(parts[1])
				if dns != "" {
					tempDNS = append(tempDNS, dns)
				}
			}

			for scanner.Scan() {
				nextLine := strings.TrimSpace(scanner.Text())
				if nextLine == "" || strings.Contains(nextLine, ":") {
					break
				}
				tempDNS = append(tempDNS, nextLine)
			}
		}
	}

	if !captured && tempIP != "" && !isLoopback {
		config.IPMethod = tempMethod
		config.IPAddress = tempIP
		config.Gateway = tempGW
		config.Subnet = tempSubnet
		config.DNS = strings.Join(tempDNS, ", ")
	}

	ifaceCmd := exec.Command("powershell", "-Command", "Get-NetAdapter | Select-Object Name, Status | ConvertTo-Json")
	ifaceOutput, err := ifaceCmd.Output()
	if err == nil {
		var adapters []NetAdapter
		err = json.Unmarshal(ifaceOutput, &adapters)
		if err != nil {
			var single NetAdapter
			if err2 := json.Unmarshal(ifaceOutput, &single); err2 == nil {
				adapters = append(adapters, single)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{
					"status":  "failed",
					"message": "Failed to parse adapter list: " + err.Error(),
				})
				return
			}
		}

		for i, adapter := range adapters {
			status := strings.ToLower(adapter.Status)
			if status == "up" || status == "connected" {
				status = "active"
			} else {
				status = "inactive"
			}
			key := strconv.Itoa(i + 1)
			config.Interface[key] = InterfaceStatus{
				Mode:   adapter.Name,
				Status: status,
			}
		}
	}

	json.NewEncoder(w).Encode(config)
}
