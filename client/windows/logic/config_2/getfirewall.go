package config_2

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
)

// Raw PowerShell rule structure
type FirewallRule struct {
	Name          string `json:"Name"`
	DisplayName   string `json:"DisplayName"`
	Direction     string `json:"Direction"`
	Action        string `json:"Action"`
	Enabled       string `json:"Enabled"`
	Profile       string `json:"Profile"`
	Protocol      string `json:"Protocol"`
	LocalPort     string `json:"LocalPort"`
	RemotePort    string `json:"RemotePort"`
	LocalAddress  string `json:"LocalAddress"`
	RemoteAddress string `json:"RemoteAddress"`
}

// Custom formatted JSON output (iptables-style)
type FirewallRuleFormatted struct {
	Type   string  `json:"type"`
	Chains []Chain `json:"chains"`
}

type Chain struct {
	Name   string `json:"name"`
	Policy string `json:"policy"`
	Rules  []Rule `json:"rules"`
}

type Rule struct {
	Chain       string `json:"chain"`
	Number      int    `json:"number"`
	Target      string `json:"target"`
	Protocol    string `json:"protocol"`
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Options     string `json:"options"`
}

func GetWindowsFirewallRules(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "failed",
			"message": "Method Not Allowed",
		})
		return
	}

	psCommand := `
Get-NetFirewallRule |
Select-Object `
	psCommand += `@{Name='Name'; Expression={$_.Name}},
@{Name='DisplayName'; Expression={$_.DisplayName}},
@{Name='Direction'; Expression={$_.Direction.ToString()}},
@{Name='Action'; Expression={$_.Action.ToString()}},
@{Name='Enabled'; Expression={$_.Enabled.ToString()}},
@{Name='Profile'; Expression={$_.Profile.ToString()}},
@{Name='Protocol'; Expression={$_.Protocol.ToString()}},
@{Name='LocalPort'; Expression={$_.LocalPort}},
@{Name='RemotePort'; Expression={$_.RemotePort}},
@{Name='LocalAddress'; Expression={$_.LocalAddress}},
@{Name='RemoteAddress'; Expression={$_.RemoteAddress}} |
ConvertTo-Json -Depth 3
`

	cmd := exec.Command("powershell", "-Command", psCommand)
	output, err := cmd.Output()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "failed",
			"message": fmt.Sprintf("Failed to execute PowerShell: %v", err),
		})
		return
	}

	var rules []FirewallRule
	err = json.Unmarshal(output, &rules)

	// Handle single object fallback
	if err != nil {
		var single FirewallRule
		if err2 := json.Unmarshal(output, &single); err2 == nil {
			rules = []FirewallRule{single}
			err = nil
		}
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "failed",
			"message": fmt.Sprintf("Failed to parse JSON output: %v", err),
		})
		return
	}

	formatted := FirewallRuleFormatted{
		Type: "firewall",
		Chains: []Chain{
			{
				Name:   "WINDOWS",
				Policy: "N/A",
				Rules:  []Rule{},
			},
		},
	}

	for i, r := range rules {
		formatted.Chains[0].Rules = append(formatted.Chains[0].Rules, Rule{
			Chain:       "WINDOWS",
			Number:      i + 1,
			Target:      r.Action,
			Protocol:    r.Protocol,
			Source:      r.RemoteAddress,
			Destination: r.LocalAddress,
			Options:     fmt.Sprintf("-- %s %s", r.RemoteAddress, r.LocalAddress),
		})
	}

	json.NewEncoder(w).Encode(formatted)
}
