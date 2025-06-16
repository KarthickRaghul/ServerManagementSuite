package config_2

import (
	"net/http"
	"os/exec"
    "encoding/json"
)

func RouteHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "failed",
			"message": "Only GET method allowed",
		})
		return
	}

	psCmd := `
$routes = Get-NetRoute |
Where-Object { $_.AddressFamily -eq 'IPv4' } |
ForEach-Object {
    $destinationPrefixParts = $_.DestinationPrefix -split '/'
    $destination = $destinationPrefixParts[0]
    $prefix = $destinationPrefixParts[1]

    try {
        $maskInt = ([int64]0xFFFFFFFF) -shl (32 - [int]$prefix)
        $genmask = [IPAddress]($maskInt -band 0xFFFFFFFF).ToString()
    } catch {
        $genmask = "0.0.0.0"
    }

    @{
        destination = $destination
        genmask     = $genmask
        gateway     = $_.NextHop
        flags       = if ($_.NextHop -ne '0.0.0.0') { "GU" } else { "U" }
        metric      = "$($_.RouteMetric)"
        ref         = "0"
        use         = "0"
        iface       = $_.InterfaceAlias
    }
}
$routes | ConvertTo-Json -Depth 3
`

	cmd := exec.Command("powershell", "-Command", psCmd)
	output, err := cmd.Output()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "failed",
			"message": "Failed to execute PowerShell command: " + err.Error(),
		})
		return
	}

	w.Write(output)
}
