package config_2

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

// RouteEntry represents a single entry in the routing table
type RouteEntry struct {
	Destination string `json:"destination"`
	Gateway     string `json:"gateway"`
	Genmask     string `json:"genmask"`
	Flags       string `json:"flags"`
	Metric      string `json:"metric"`
	Ref         string `json:"ref"`
	Use         string `json:"use"`
	Iface       string `json:"iface"` // Now shows interface name (e.g., Ethernet)
}

// HandleRouteTable handles requests to get the routing table on Windows
func HandleRouteTable(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Current Date and Time (UTC - YYYY-MM-DD HH:MM:SS formatted):",
		time.Now().UTC().Format("2006-01-02 15:04:05"))
	fmt.Println("Current User's Login: kishore-001")
	fmt.Println("Handling route table request...")

	if r.Method != http.MethodGet {
		http.Error(w, "Only GET allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	routes, err := getWindowsRoutingTable()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error":   "Failed to get routing table",
			"details": err.Error(),
		})
		return
	}

	fmt.Println("Sending route table response...")
	json.NewEncoder(w).Encode(routes)
}

// getWindowsRoutingTable parses output from 'route print' on Windows and normalizes iface names
func getWindowsRoutingTable() ([]RouteEntry, error) {
	out, err := exec.Command("route", "print").Output()
	if err != nil {
		return nil, fmt.Errorf("route print failed: %w", err)
	}

	ifaceMap, _ := getInterfaceNameMap()

	lines := strings.Split(string(out), "\n")
	routes := []RouteEntry{}
	inIPv4Section := false

	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])

		if strings.Contains(line, "IPv4 Route Table") {
			inIPv4Section = true
			continue
		}
		if inIPv4Section && strings.Contains(line, "Active Routes:") {
			i += 2 // skip header and column line
			for ; i < len(lines); i++ {
				line = strings.TrimSpace(lines[i])
				if line == "" || strings.Contains(line, "Persistent Routes") {
					break
				}

				fields := splitByWhitespace(line)
				if len(fields) < 5 {
					continue
				}

				ifaceIP := fields[3]
				ifaceName := ifaceIP
				if name, ok := ifaceMap[ifaceIP]; ok {
					ifaceName = name
				}

				route := RouteEntry{
					Destination: fields[0],
					Genmask:     fields[1],
					Gateway:     fields[2],
					Iface:       ifaceName,
					Metric:      fields[4],
					Flags:       "U",
					Ref:         "0",
					Use:         "0",
				}

				if route.Gateway != "0.0.0.0" && route.Gateway != "On-link" {
					route.Flags += "G"
				}

				routes = append(routes, route)
			}
			break
		}
	}

	return routes, nil
}

// splitByWhitespace splits a string by multiple spaces
func splitByWhitespace(s string) []string {
	re := regexp.MustCompile(`\s+`)
	return re.Split(strings.TrimSpace(s), -1)
}

// getInterfaceNameMap returns a map[ipAddress] = interfaceName from PowerShell
func getInterfaceNameMap() (map[string]string, error) {
	cmd := exec.Command("powershell", "-Command", "Get-NetIPAddress | Select-Object -Property IPAddress, InterfaceAlias | ConvertTo-Csv -NoTypeInformation")
	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		return nil, err
	}

	reader := csv.NewReader(&out)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	ifaceMap := make(map[string]string)
	for _, record := range records[1:] { // Skip header
		if len(record) >= 2 {
			ip := strings.TrimSpace(record[0])
			iface := strings.TrimSpace(record[1])
			if ip != "" && iface != "" {
				ifaceMap[ip] = iface
			}
		}
	}
	return ifaceMap, nil
}
