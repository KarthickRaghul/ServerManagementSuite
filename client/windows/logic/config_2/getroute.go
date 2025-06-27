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

// getWindowsRoutingTable parses output from 'route print' on Windows
func getWindowsRoutingTable() ([]RouteEntry, error) {
	out, err := exec.Command("route", "print").Output()
	if err != nil {
		return nil, fmt.Errorf("route print failed: %w", err)
	}

	lines := strings.Split(string(out), "\n")
	routes := []RouteEntry{}

	// Start parsing from "IPv4 Route Table"
	inIPv4Section := false
	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])

		// Detect start of IPv4 table
		if strings.Contains(line, "IPv4 Route Table") {
			inIPv4Section = true
			continue
		}
		if inIPv4Section && strings.Contains(line, "Active Routes:") {
			// Skip header line and column headers
			i += 2
			for ; i < len(lines); i++ {
				line = strings.TrimSpace(lines[i])
				if line == "" || strings.Contains(line, "Persistent Routes") {
					break
				}

				fields := splitByWhitespace(line)
				if len(fields) < 5 {
					continue
				}

				// Build RouteEntry with default values for missing fields
				route := RouteEntry{
					Destination: fields[0],
					Genmask:     fields[1],
					Gateway:     fields[2],
					Iface:       fields[3],
					Metric:      fields[4],
					Flags:       "U",
					Ref:         "0",
					Use:         "0",
				}

				// Set flag G for Gateway 0.0.0.0 or non-direct
				if route.Gateway != "0.0.0.0" {
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
