package config_2

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// FirewallRule represents a single firewall rule
type FirewallRule struct {
	Chain       string `json:"chain"`
	Number      int    `json:"number"`
	Target      string `json:"target"`
	Protocol    string `json:"protocol"`
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Interface   string `json:"interface,omitempty"`
	State       string `json:"state,omitempty"`
	Ports       string `json:"ports,omitempty"`
	Options     string `json:"options,omitempty"`
}

// FirewallChain represents a chain of firewall rules
type FirewallChain struct {
	Name   string         `json:"name"`
	Policy string         `json:"policy"`
	Rules  []FirewallRule `json:"rules"`
}

// FirewallResponse represents the overall firewall configuration
type FirewallResponse struct {
	Type   string          `json:"type"` // "iptables", "ufw", "firewalld"
	Chains []FirewallChain `json:"chains"`
	Active bool            `json:"active"`
}

// HandleFirewallRules handles requests to get the firewall rules
func HandleFirewallRules(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Current Date and Time (UTC - YYYY-MM-DD HH:MM:SS formatted):",
		time.Now().UTC().Format("2006-01-02 15:04:05"))
	fmt.Println("Current User's Login: kishore-001")
	fmt.Println("Handling firewall rules request...")

	// Check for GET method
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET allowed", http.StatusMethodNotAllowed)
		return
	}

	// Set content type
	w.Header().Set("Content-Type", "application/json")

	// Try to get firewall rules using different methods
	firewallRules, err := getFirewallRules()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		errorResp := map[string]string{
			"error":   "Failed to get firewall rules",
			"details": err.Error(),
		}
		json.NewEncoder(w).Encode(errorResp)
		return
	}

	fmt.Println("Sending firewall rules response...")
	// Send the firewall rules using NewEncoder().Encode()
	json.NewEncoder(w).Encode(firewallRules)
}

// getFirewallRules retrieves firewall rules from the system
func getFirewallRules() (*FirewallResponse, error) {
	// Check which firewall system is active
	response := &FirewallResponse{
		Chains: []FirewallChain{},
		Active: false,
	}

	// Try UFW first (Ubuntu/Debian default)
	ufwRules, ufwActive, ufwErr := getUFWRules()
	if ufwErr == nil && ufwActive {
		response.Type = "ufw"
		response.Chains = ufwRules
		response.Active = true
		return response, nil
	}

	// Try FirewallD (RHEL/CentOS/Fedora default)
	firewallDRules, firewallDActive, firewallDErr := getFirewallDRules()
	if firewallDErr == nil && firewallDActive {
		response.Type = "firewalld"
		response.Chains = firewallDRules
		response.Active = true
		return response, nil
	}

	// Fall back to plain iptables
	iptablesRules, iptablesErr := getIPTablesRules()
	if iptablesErr == nil {
		response.Type = "iptables"
		response.Chains = iptablesRules
		response.Active = true
		return response, nil
	}

	// If all methods failed, check if any firewalls are installed but inactive
	if ufwErr == nil {
		response.Type = "ufw"
		response.Chains = ufwRules
		response.Active = false
		return response, nil
	}

	if firewallDErr == nil {
		response.Type = "firewalld"
		response.Chains = firewallDRules
		response.Active = false
		return response, nil
	}

	// If we can't find any firewall information
	if ufwErr != nil && firewallDErr != nil && iptablesErr != nil {
		return nil, fmt.Errorf("no firewall system found or accessible")
	}

	return response, nil
}

// getIPTablesRules gets rules directly from iptables command
func getIPTablesRules() ([]FirewallChain, error) {
	chains := []FirewallChain{}

	// Check if iptables is installed
	_, err := exec.LookPath("iptables")
	if err != nil {
		return nil, fmt.Errorf("iptables not found: %w", err)
	}

	// Get list of chains
	cmd := exec.Command("iptables", "-L", "-n", "--line-numbers")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run iptables command: %w", err)
	}

	// Parse the output
	outputStr := string(output)
	chainBlocks := strings.Split(outputStr, "Chain ")

	// Skip the first empty block
	for _, block := range chainBlocks[1:] {
		lines := strings.Split(block, "\n")
		if len(lines) < 2 {
			continue
		}

		// Parse chain name and policy
		chainHeader := lines[0]
		headerParts := strings.Fields(chainHeader)
		if len(headerParts) < 2 {
			continue
		}

		chain := FirewallChain{
			Name:   headerParts[0],
			Policy: "ACCEPT", // Default policy
			Rules:  []FirewallRule{},
		}

		// Extract policy if specified
		if strings.Contains(chainHeader, "policy") {
			policyMatch := regexp.MustCompile(`\(policy\s+(\w+)\)`).FindStringSubmatch(chainHeader)
			if len(policyMatch) > 1 {
				chain.Policy = policyMatch[1]
			}
		}

		// Skip the headers (first 2 lines of each block)
		for _, line := range lines[2:] {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			// Parse rule
			fields := strings.Fields(line)
			if len(fields) < 4 {
				continue
			}

			num, err := strconv.Atoi(fields[0])
			if err != nil {
				continue // Skip if number can't be parsed
			}

			rule := FirewallRule{
				Chain:    chain.Name,
				Number:   num,
				Target:   fields[1],
				Protocol: fields[2],
				Options:  "",
			}

			// Extract source and destination
			for i, field := range fields {
				if i > 2 {
					if field == "anywhere" {
						if rule.Source == "" {
							rule.Source = field
						} else {
							rule.Destination = field
						}
					} else if strings.Contains(field, "/") || strings.Contains(field, ".") {
						if rule.Source == "" {
							rule.Source = field
						} else {
							rule.Destination = field
						}
					}
				}
			}

			// Extract interface if specified
			ifaceInMatch := regexp.MustCompile(`in:\s*(\S+)`).FindStringSubmatch(line)
			ifaceOutMatch := regexp.MustCompile(`out:\s*(\S+)`).FindStringSubmatch(line)
			if len(ifaceInMatch) > 1 {
				rule.Interface = "in:" + ifaceInMatch[1]
			} else if len(ifaceOutMatch) > 1 {
				rule.Interface = "out:" + ifaceOutMatch[1]
			}

			// Extract state if specified
			stateMatch := regexp.MustCompile(`state\s+(\S+)`).FindStringSubmatch(line)
			if len(stateMatch) > 1 {
				rule.State = stateMatch[1]
			}

			// Extract ports if specified
			dportMatch := regexp.MustCompile(`dpt:(\S+)`).FindStringSubmatch(line)
			sportMatch := regexp.MustCompile(`spt:(\S+)`).FindStringSubmatch(line)
			if len(dportMatch) > 1 {
				rule.Ports = "dpt:" + dportMatch[1]
			} else if len(sportMatch) > 1 {
				rule.Ports = "spt:" + sportMatch[1]
			}

			// Anything else goes in options
			rule.Options = strings.Join(fields[3:], " ")

			chain.Rules = append(chain.Rules, rule)
		}

		chains = append(chains, chain)
	}

	return chains, nil
}

// getUFWRules gets rules from ufw (Uncomplicated Firewall)
func getUFWRules() ([]FirewallChain, bool, error) {
	chains := []FirewallChain{}

	// Check if ufw is installed
	_, err := exec.LookPath("ufw")
	if err != nil {
		return nil, false, fmt.Errorf("ufw not found: %w", err)
	}

	// Check if ufw is enabled
	statusCmd := exec.Command("ufw", "status")
	statusOutput, err := statusCmd.Output()
	if err != nil {
		return nil, false, fmt.Errorf("failed to run ufw status command: %w", err)
	}

	isActive := strings.Contains(string(statusOutput), "Status: active")

	// Get verbose output for parsing
	cmd := exec.Command("ufw", "status", "verbose", "numbered")
	output, err := cmd.Output()
	if err != nil {
		// If we can't get the verbose output, try the normal one
		cmd = exec.Command("ufw", "status", "numbered")
		output, err = cmd.Output()
		if err != nil {
			return nil, isActive, fmt.Errorf("failed to run ufw status numbered command: %w", err)
		}
	}

	// Parse the output
	outputStr := string(output)
	lines := strings.Split(outputStr, "\n")

	// Create a single chain for UFW rules
	chain := FirewallChain{
		Name:   "UFW",
		Policy: "DROP", // Default UFW policy
		Rules:  []FirewallRule{},
	}

	// Extract the default policy if available
	for _, line := range lines {
		if strings.Contains(line, "Default:") {
			policyMatch := regexp.MustCompile(`Default:\s+(\w+)`).FindStringSubmatch(line)
			if len(policyMatch) > 1 {
				chain.Policy = policyMatch[1]
			}
		}
	}

	// Find the rule section (starts after the To Action From line)
	inRuleSection := false
	ruleCount := 1
	for _, line := range lines {
		if strings.Contains(line, "To") && strings.Contains(line, "Action") && strings.Contains(line, "From") {
			inRuleSection = true
			continue
		}

		if inRuleSection && line != "" {
			// Extract numbered rule if possible
			numMatch := regexp.MustCompile(`\[(\d+)\]`).FindStringSubmatch(line)
			num := ruleCount
			if len(numMatch) > 1 {
				parsedNum, err := strconv.Atoi(numMatch[1])
				if err == nil {
					num = parsedNum
				}
			}

			// Parse rule fields
			fields := strings.Fields(line)
			if len(fields) < 3 {
				continue
			}

			// Try to extract the relevant parts
			var action, from, to, ports string
			for i, field := range fields {
				if field == "ALLOW" || field == "DENY" || field == "REJECT" || field == "LIMIT" {
					action = field
				} else if strings.Contains(field, "Anywhere") || strings.Contains(field, "/") {
					if from == "" {
						from = field
					} else {
						to = field
					}
				} else if strings.Contains(field, "tcp") || strings.Contains(field, "udp") {
					ports = strings.Join(fields[i:i+2], " ")
					break
				}
			}

			rule := FirewallRule{
				Chain:       "UFW",
				Number:      num,
				Target:      action,
				Protocol:    "all", // Default
				Source:      from,
				Destination: to,
				Ports:       ports,
				Options:     strings.TrimSpace(line),
			}

			chain.Rules = append(chain.Rules, rule)
			ruleCount++
		}
	}

	chains = append(chains, chain)
	return chains, isActive, nil
}

// getFirewallDRules gets rules from firewalld (RHEL/CentOS/Fedora firewall)
func getFirewallDRules() ([]FirewallChain, bool, error) {
	chains := []FirewallChain{}

	// Check if firewall-cmd is installed
	_, err := exec.LookPath("firewall-cmd")
	if err != nil {
		return nil, false, fmt.Errorf("firewall-cmd not found: %w", err)
	}

	// Check if firewalld is running
	statusCmd := exec.Command("firewall-cmd", "--state")
	_, err = statusCmd.Output()
	isActive := err == nil

	// Get default zone
	defaultZoneCmd := exec.Command("firewall-cmd", "--get-default-zone")
	defaultZoneOutput, err := defaultZoneCmd.Output()
	if err != nil {
		return nil, isActive, fmt.Errorf("failed to get default zone: %w", err)
	}
	defaultZone := strings.TrimSpace(string(defaultZoneOutput))

	// Get zones
	zonesCmd := exec.Command("firewall-cmd", "--get-zones")
	zonesOutput, err := zonesCmd.Output()
	if err != nil {
		return nil, isActive, fmt.Errorf("failed to get zones: %w", err)
	}
	zones := strings.Fields(string(zonesOutput))

	// Process each zone
	for _, zone := range zones {
		chain := FirewallChain{
			Name:   zone,
			Policy: "reject", // Default firewalld policy
			Rules:  []FirewallRule{},
		}

		if zone == defaultZone {
			chain.Name = zone + " (default)"
		}

		// Get services
		servicesCmd := exec.Command("firewall-cmd", "--zone="+zone, "--list-services")
		servicesOutput, err := servicesCmd.Output()
		if err == nil {
			services := strings.Fields(string(servicesOutput))
			for i, service := range services {
				rule := FirewallRule{
					Chain:       zone,
					Number:      i + 1,
					Target:      "ACCEPT",
					Protocol:    "all",
					Source:      "any",
					Destination: "any",
					Options:     "service: " + service,
				}
				chain.Rules = append(chain.Rules, rule)
			}
		}

		// Get ports
		portsCmd := exec.Command("firewall-cmd", "--zone="+zone, "--list-ports")
		portsOutput, err := portsCmd.Output()
		if err == nil {
			ports := strings.Fields(string(portsOutput))
			offset := len(chain.Rules)
			for i, port := range ports {
				protocol := "tcp"
				if strings.Contains(port, "udp") {
					protocol = "udp"
				}

				rule := FirewallRule{
					Chain:       zone,
					Number:      offset + i + 1,
					Target:      "ACCEPT",
					Protocol:    protocol,
					Source:      "any",
					Destination: "any",
					Ports:       port,
				}
				chain.Rules = append(chain.Rules, rule)
			}
		}

		// Get rich rules
		richRulesCmd := exec.Command("firewall-cmd", "--zone="+zone, "--list-rich-rules")
		richRulesOutput, err := richRulesCmd.Output()
		if err == nil {
			richRulesStr := string(richRulesOutput)
			richRules := strings.Split(richRulesStr, "\n")
			offset := len(chain.Rules)

			for i, richRule := range richRules {
				richRule = strings.TrimSpace(richRule)
				if richRule == "" {
					continue
				}

				rule := FirewallRule{
					Chain:       zone,
					Number:      offset + i + 1,
					Target:      "RICH-RULE",
					Protocol:    "all",
					Source:      "any",
					Destination: "any",
					Options:     richRule,
				}

				// Try to extract more details
				if strings.Contains(richRule, "service name=") {
					serviceMatch := regexp.MustCompile(`service name="([^"]+)"`).FindStringSubmatch(richRule)
					if len(serviceMatch) > 1 {
						rule.Options = "service: " + serviceMatch[1]
					}
				} else if strings.Contains(richRule, "port port=") {
					portMatch := regexp.MustCompile(`port port="(\d+)"\s+protocol="([^"]+)"`).FindStringSubmatch(richRule)
					if len(portMatch) > 2 {
						rule.Ports = portMatch[1] + "/" + portMatch[2]
						rule.Protocol = portMatch[2]
					}
				}

				// Extract action
				if strings.Contains(richRule, "accept") {
					rule.Target = "ACCEPT"
				} else if strings.Contains(richRule, "reject") {
					rule.Target = "REJECT"
				} else if strings.Contains(richRule, "drop") {
					rule.Target = "DROP"
				}

				chain.Rules = append(chain.Rules, rule)
			}
		}

		chains = append(chains, chain)
	}

	return chains, isActive, nil
}
