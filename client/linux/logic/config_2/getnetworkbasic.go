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
	Power  string `json:"power"`
}

// NetworkConfigResponse represents the response format you specified
type NetworkConfigResponse struct {
	IPMethod  string                   `json:"ip_method"`
	IPAddress string                   `json:"ip_address"`
	Gateway   string                   `json:"gateway"`
	Subnet    string                   `json:"subnet"`
	DNS       string                   `json:"dns"`
	Uptime    string                   `json:"uptime"`
	Interface map[string]InterfaceInfo `json:"interface"`
}

// ErrorResponse represents the standard error response format


// HandleNetworkConfig provides network configuration as an HTTP response
// in the exact format specified
func HandleNetworkConfig(w http.ResponseWriter, r *http.Request) {

	// Check for GET method
	if r.Method != http.MethodGet {
		sendError(w, "Only GET method allowed", http.StatusMethodNotAllowed)
		return
	}

	// Set content type
	w.Header().Set("Content-Type", "application/json")

	// Get all network interfaces
	interfaces, err := net.Interfaces()
	if err != nil {
		sendError(w, "Failed to get network interfaces: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Get default route to determine active interface, gateway and IP method
	defaultRoutes, err := getDefaultRoutes()
	if err != nil {
		defaultRoutes = []DefaultRoute{} // Continue with empty route info
	}

	// Find primary active interface (the one with lowest metric)
	activeRoute := getActiveRoute(defaultRoutes)

	// Get DNS servers
	dnsServers, err := getDNSServers()
	if err != nil {
		dnsServers = []string{} // Continue with empty DNS info
	}

	// Get Uptime
	uptime, err := getSystemUptime()
	if err != nil {
		uptime = "unknown"
	}

	// Prepare response in the requested format
	response := NetworkConfigResponse{
		IPMethod:  "static", // Default to static
		IPAddress: "",
		Gateway:   activeRoute.Gateway,
		Subnet:    "",
		DNS:       strings.Join(dnsServers, ", "),
		Uptime:    uptime,
		Interface: make(map[string]InterfaceInfo),
	}

	// Set IP method based on active route's proto
	if activeRoute.Proto == "dhcp" {
		response.IPMethod = "dynamic"
	}

	// Process each interface
	interfaceCount := 1
	for _, iface := range interfaces {
		// Skip loopback and virtual interfaces
		if iface.Flags&net.FlagLoopback != 0 ||
			strings.Contains(iface.Name, "docker") ||
			strings.Contains(iface.Name, "veth") ||
			strings.Contains(iface.Name, "br-") {
			continue
		}

		// Determine power status (up/down)
		power := "off"
		if iface.Flags&net.FlagUp != 0 {
			power = "on"
		}

		// Determine if this interface is the active one
		status := "inactive"
		if iface.Name == activeRoute.Interface {
			status = "active"
		}

		// Add interface to the map
		response.Interface[fmt.Sprintf("%d", interfaceCount)] = InterfaceInfo{
			Mode:   iface.Name,
			Status: status,
			Power:  power,
		}
		interfaceCount++

		// Get IP address and subnet for active interface
		if status == "active" {
			// Use IP address from the route if available
			if activeRoute.SourceIP != "" {
				response.IPAddress = activeRoute.SourceIP
			} else {
				// Fall back to getting IP from interface if not in route info
				addrs, err := iface.Addrs()
				if err == nil {
					for _, addr := range addrs {
						if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
							if ipnet.IP.To4() != nil { // Only consider IPv4 for now
								response.IPAddress = ipnet.IP.String()
								break
							}
						}
					}
				}
			}

			// Get subnet information
			response.Subnet = getSubnetForInterface(iface.Name)
		}
	}

	fmt.Println("Sending network configuration response...")
	// Send the network configuration using NewEncoder().Encode()
	json.NewEncoder(w).Encode(response)
}

// DefaultRoute represents a route with additional information
type DefaultRoute struct {
	Interface string
	Gateway   string
	Proto     string // dhcp, static, etc.
	SourceIP  string // Source IP address
	Metric    int    // Route priority
}

// getDefaultRoutes gets all default gateway routes
func getDefaultRoutes() ([]DefaultRoute, error) {
	routes := []DefaultRoute{}

	// Run ip route command to get default routes
	cmd := exec.Command("ip", "route", "show", "default")
	output, err := cmd.Output()
	if err != nil {
		return routes, err
	}

	// Parse each default route line
	routeStr := string(output)
	if routeStr == "" {
		return routes, fmt.Errorf("no default route found")
	}

	for _, line := range strings.Split(routeStr, "\n") {
		if line == "" {
			continue
		}

		// Skip non-default routes
		if !strings.HasPrefix(line, "default") {
			continue
		}

		route := DefaultRoute{
			Metric: 999999, // Default high metric
		}

		parts := strings.Fields(line)
		for i, part := range parts {
			switch part {
			case "via":
				if i+1 < len(parts) {
					route.Gateway = parts[i+1]
				}
			case "dev":
				if i+1 < len(parts) {
					route.Interface = parts[i+1]
				}
			case "proto":
				if i+1 < len(parts) {
					route.Proto = parts[i+1]
				}
			case "src":
				if i+1 < len(parts) {
					route.SourceIP = parts[i+1]
				}
			case "metric":
				if i+1 < len(parts) {
					fmt.Sscanf(parts[i+1], "%d", &route.Metric)
				}
			}
		}

		// Only add routes with an interface
		if route.Interface != "" {
			routes = append(routes, route)
		}
	}

	return routes, nil
}

// getActiveRoute returns the route with the lowest metric (highest priority)
func getActiveRoute(routes []DefaultRoute) DefaultRoute {
	if len(routes) == 0 {
		return DefaultRoute{}
	}

	activeRoute := routes[0]
	for _, route := range routes {
		if route.Metric < activeRoute.Metric {
			activeRoute = route
		}
	}

	return activeRoute
}

// getSubnetForInterface gets the subnet for a specific interface
func getSubnetForInterface(ifaceName string) string {
	// Run ip addr command to get subnet information
	cmd := exec.Command("ip", "addr", "show", ifaceName)
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	// Parse output for IPv4 subnet in CIDR notation
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "inet ") {
			parts := strings.Fields(line)
			for _, part := range parts {
				if strings.Contains(part, "/") && net.ParseIP(strings.Split(part, "/")[0]).To4() != nil {
					return part // Return the CIDR notation (e.g., 192.168.1.100/24)
				}
			}
		}
	}

	return ""
}

// getDNSServers reads DNS servers from /etc/resolv.conf
func getDNSServers() ([]string, error) {
	servers := []string{}

	// Run grep to extract nameservers from resolv.conf
	cmd := exec.Command("grep", "^nameserver", "/etc/resolv.conf")
	output, err := cmd.Output()
	if err != nil {
		return servers, err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			servers = append(servers, parts[1])
		}
	}

	return servers, nil
}

// obtain the uptime for the server
func getSystemUptime() (string, error) {
	// This works on Linux; reads uptime in seconds from /proc/uptime
	data, err := exec.Command("cut", "-d", " ", "-f1", "/proc/uptime").Output()
	if err != nil {
		return "", err
	}

	// Parse float value and convert to human readable format
	secondsStr := strings.TrimSpace(string(data))
	var seconds float64
	_, err = fmt.Sscanf(secondsStr, "%f", &seconds)
	if err != nil {
		return "", err
	}

	dur := int64(seconds)
	days := dur / (60 * 60 * 24)
	hours := (dur / (60 * 60)) % 24
	minutes := (dur / 60) % 60
	secs := dur % 60

	return fmt.Sprintf("%dd %dh %dm %ds", days, hours, minutes, secs), nil
}

