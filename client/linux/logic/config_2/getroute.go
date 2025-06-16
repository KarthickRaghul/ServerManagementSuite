package config_2

import (
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
	Iface       string `json:"iface"`
}

// HandleRouteTable handles requests to get the routing table
func HandleRouteTable(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Current Date and Time (UTC - YYYY-MM-DD HH:MM:SS formatted):",
		time.Now().UTC().Format("2006-01-02 15:04:05"))
	fmt.Println("Current User's Login: kishore-001")
	fmt.Println("Handling route table request...")

	// Check for GET method
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET allowed", http.StatusMethodNotAllowed)
		return
	}

	// Set content type
	w.Header().Set("Content-Type", "application/json")

	// Get the routing table
	routes, err := getRoutingTable()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		errorResp := map[string]string{
			"error":   "Failed to get routing table",
			"details": err.Error(),
		}
		json.NewEncoder(w).Encode(errorResp)
		return
	}

	fmt.Println("Sending route table response...")
	// Send the routing table using NewEncoder().Encode()
	json.NewEncoder(w).Encode(routes)
}

// getRoutingTable retrieves the routing table from the system
func getRoutingTable() ([]RouteEntry, error) {
	routes := []RouteEntry{}

	// Try the 'ip route' command first (more modern)
	ipRoutes, ipErr := getIpRoutes()
	if ipErr == nil {
		return ipRoutes, nil
	}

	// Fall back to 'route' command if 'ip route' fails
	routeCmd := exec.Command("route", "-n")
	output, err := routeCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run route command: %w", err)
	}

	// Parse the output
	lines := strings.Split(string(output), "\n")
	if len(lines) < 3 {
		return nil, fmt.Errorf("unexpected route command output format")
	}

	// Skip the first two lines (headers)
	for _, line := range lines[2:] {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Split the line by whitespace
		fields := splitByWhitespace(line)
		if len(fields) < 8 {
			continue // Skip invalid lines
		}

		route := RouteEntry{
			Destination: fields[0],
			Gateway:     fields[1],
			Genmask:     fields[2],
			Flags:       fields[3],
			Metric:      fields[4],
			Ref:         fields[5],
			Use:         fields[6],
			Iface:       fields[7],
		}

		routes = append(routes, route)
	}

	return routes, nil
}

// getIpRoutes tries to get routes using the 'ip route' command
func getIpRoutes() ([]RouteEntry, error) {
	routes := []RouteEntry{}

	// Run ip route command
	cmd := exec.Command("ip", "route")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	// Parse the output
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse the ip route line
		route := parseIpRouteLine(line)
		routes = append(routes, route)
	}

	return routes, nil
}

// parseIpRouteLine parses a single line from 'ip route' output
func parseIpRouteLine(line string) RouteEntry {
	// Initialize with default values
	route := RouteEntry{
		Destination: "",
		Gateway:     "",
		Genmask:     "",
		Flags:       "",
		Metric:      "",
		Ref:         "",
		Use:         "",
		Iface:       "",
	}

	// Extract destination
	if strings.HasPrefix(line, "default") {
		route.Destination = "0.0.0.0"
		route.Genmask = "0.0.0.0"
	} else {
		destParts := strings.Fields(line)
		if len(destParts) > 0 {
			// Handle CIDR notation like 192.168.1.0/24
			if strings.Contains(destParts[0], "/") {
				cidrParts := strings.Split(destParts[0], "/")
				if len(cidrParts) == 2 {
					route.Destination = cidrParts[0]
					// Convert CIDR prefix to netmask (simplified)
					route.Genmask = cidrToNetmask(cidrParts[1])
				}
			} else {
				route.Destination = destParts[0]
			}
		}
	}

	// Extract gateway
	gatewayMatch := regexp.MustCompile(`via\s+(\S+)`).FindStringSubmatch(line)
	if len(gatewayMatch) > 1 {
		route.Gateway = gatewayMatch[1]
	} else {
		route.Gateway = "*"
	}

	// Extract interface
	ifaceMatch := regexp.MustCompile(`dev\s+(\S+)`).FindStringSubmatch(line)
	if len(ifaceMatch) > 1 {
		route.Iface = ifaceMatch[1]
	}

	// Extract metric
	metricMatch := regexp.MustCompile(`metric\s+(\d+)`).FindStringSubmatch(line)
	if len(metricMatch) > 1 {
		route.Metric = metricMatch[1]
	} else {
		route.Metric = "0"
	}

	// Set flags based on keywords in the line
	flags := ""
	if strings.Contains(line, "default") {
		flags += "G"
	}
	if strings.Contains(line, "linkdown") {
		flags += "!"
	} else {
		flags += "U"
	}
	route.Flags = flags

	route.Ref = "0" // Always 0 for ip route output
	route.Use = "0" // Always 0 for ip route output

	return route
}

// cidrToNetmask converts a CIDR prefix length to a netmask string
func cidrToNetmask(cidr string) string {
	prefixLen := 0
	fmt.Sscanf(cidr, "%d", &prefixLen)

	if prefixLen < 0 || prefixLen > 32 {
		return "255.255.255.255" // Invalid, return full mask
	}

	// Calculate the netmask
	mask := uint32(0xFFFFFFFF) << (32 - prefixLen)

	// Convert to dotted decimal format
	return fmt.Sprintf("%d.%d.%d.%d",
		(mask>>24)&0xFF,
		(mask>>16)&0xFF,
		(mask>>8)&0xFF,
		mask&0xFF)
}

// splitByWhitespace splits a string by whitespace, handling multiple spaces
func splitByWhitespace(s string) []string {
	// Replace multiple whitespace with a single space
	re := regexp.MustCompile(`\s+`)
	s = re.ReplaceAllString(s, " ")
	return strings.Split(strings.TrimSpace(s), " ")
}
