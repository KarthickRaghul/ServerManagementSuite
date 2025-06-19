package config_2

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
)

// MinimalFirewallRule represents only the fast-access metadata
type MinimalFirewallRule struct {
	Name        string `json:"Name"`
	DisplayName string `json:"DisplayName"`
	Direction   string `json:"Direction"`
	Action      string `json:"Action"`
	Enabled     string `json:"Enabled"`
	Profile     string `json:"Profile"`
}

// GetWindowsFirewallRulesFast returns basic firewall rules (fast, no filtering)
func GetWindowsFirewallRulesFast(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "failed",
			"message": "Method Not Allowed",
		})
		return
	}

	// PowerShell command to get basic firewall rule info quickly
	psCommand := `
	Get-NetFirewallRule | Select-Object Name,DisplayName,@{Name='Direction';Expression={ $_.Direction.ToString() }},@{Name='Action';Expression={ $_.Action.ToString() }},@{Name='Enabled';Expression={ $_.Enabled.ToString() }},@{Name='Profile';Expression={ $_.Profile.ToString() }} | ConvertTo-Json -Depth 2
`

	cmd := exec.Command("powershell", "-Command", psCommand)
	output, err := cmd.Output()
	if err != nil {
		writeError(w, "Failed to execute PowerShell", err)
		return
	}

	var rules []MinimalFirewallRule
	if err := json.Unmarshal(output, &rules); err != nil {
		// Try single object fallback
		var single MinimalFirewallRule
		if err2 := json.Unmarshal(output, &single); err2 == nil {
			rules = []MinimalFirewallRule{single}
		} else {
			writeError(w, "Failed to parse firewall rules JSON", err)
			return
		}
	}

	json.NewEncoder(w).Encode(rules)
}

func writeError(w http.ResponseWriter, message string, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "failed",
		"message": fmt.Sprintf("%s: %v", message, err),
	})
}
