package config2

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"backend/config"
	serverdb "backend/db/gen/server"
)

type firewallUpdateRequest struct {
	Host        string `json:"host"`
	Action      string `json:"action"`      // add/delete
	Rule        string `json:"rule"`        // accept/drop/reject (Linux) or allow/block (Windows)
	Protocol    string `json:"protocol"`    // tcp/udp
	Port        string `json:"port"`        // e.g., "80"
	Source      string `json:"source"`      // optional
	Destination string `json:"destination"` // optional

	// Windows-specific fields
	Name          string `json:"name,omitempty"`          // Windows rule name
	DisplayName   string `json:"displayName,omitempty"`   // Windows display name
	Direction     string `json:"direction,omitempty"`     // Inbound/Outbound
	ActionType    string `json:"actionType,omitempty"`    // Allow/Block (Windows)
	Enabled       string `json:"enabled,omitempty"`       // True/False
	Profile       string `json:"profile,omitempty"`       // Public/Private/Domain/Any
	LocalPort     string `json:"localPort,omitempty"`     // Windows local port
	RemotePort    string `json:"remotePort,omitempty"`    // Windows remote port
	LocalAddress  string `json:"localAddress,omitempty"`  // Windows local address
	RemoteAddress string `json:"remoteAddress,omitempty"` // Windows remote address
	Program       string `json:"program,omitempty"`       // Windows program path
	Service       string `json:"service,omitempty"`       // Windows service name
}

// Enhanced logging function with request ID
func logWithRequestID(requestID, level, component, message string, args ...interface{}) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	formattedMessage := fmt.Sprintf(message, args...)
	log.Printf("[%s] [%s] [%s] [%s] %s", timestamp, requestID, level, component, formattedMessage)
}

// Generate unique request ID
func generateRequestID() string {
	return fmt.Sprintf("fw-%d", time.Now().UnixNano())
}

func HandlePostUpdateFirewall(queries *serverdb.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Generate unique request ID for tracing
		requestID := generateRequestID()

		// Enhanced request logging
		logWithRequestID(requestID, "INFO", "FIREWALL", "=== FIREWALL REQUEST START ===")
		logWithRequestID(requestID, "INFO", "FIREWALL", "Method: %s, URL: %s, RemoteAddr: %s", r.Method, r.URL.Path, r.RemoteAddr)

		w.Header().Set("X-Request-ID", requestID) // Add request ID to response headers

		// Only allow POST
		if r.Method != http.MethodPost {
			logWithRequestID(requestID, "ERROR", "FIREWALL", "Method not allowed: %s", r.Method)
			sendError(w, "Only POST method allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse request body
		var req firewallUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logWithRequestID(requestID, "ERROR", "FIREWALL", "Failed to unmarshal request: %v", err)
			sendError(w, "Invalid request format: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Log parsed request structure
		logWithRequestID(requestID, "DEBUG", "FIREWALL", "Parsed request: %+v", req)

		// Basic validation
		if req.Host == "" || req.Action == "" {
			logWithRequestID(requestID, "ERROR", "FIREWALL", "Missing required fields - Host: '%s', Action: '%s'", req.Host, req.Action)
			sendError(w, "Host and action are required", http.StatusBadRequest)
			return
		}

		logWithRequestID(requestID, "INFO", "FIREWALL", "Processing firewall %s for host: %s", req.Action, req.Host)

		// Lookup device and get access token
		device, err := queries.GetServerDeviceByIP(context.Background(), req.Host)
		if err == sql.ErrNoRows {
			logWithRequestID(requestID, "ERROR", "FIREWALL", "Device not found: %s", req.Host)
			sendError(w, "Device not registered", http.StatusNotFound)
			return
		} else if err != nil {
			logWithRequestID(requestID, "ERROR", "FIREWALL", "Database error: %v", err)
			sendError(w, "Database error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		logWithRequestID(requestID, "DEBUG", "FIREWALL", "Found device: ID=%s, Tag=%s, OS=%s", device.ID, device.Tag, device.Os)

		// Process firewall request based on OS
		clientPayload := processFirewallUpdateRequest(req, device.Os, requestID)

		// Convert to JSON for client
		clientBody, err := json.Marshal(clientPayload)
		if err != nil {
			logWithRequestID(requestID, "ERROR", "FIREWALL", "Failed to marshal client payload: %v", err)
			sendError(w, "Failed to prepare client request: "+err.Error(), http.StatusInternalServerError)
			return
		}

		logWithRequestID(requestID, "DEBUG", "FIREWALL", "Client payload: %s", string(clientBody))

		// Use config for client URL
		clientURL := config.GetClientURL(req.Host, "/client/config2/updatefirewall")
		logWithRequestID(requestID, "INFO", "FIREWALL", "Sending request to: %s", clientURL)

		clientReq, err := http.NewRequest("POST", clientURL, bytes.NewReader(clientBody))
		if err != nil {
			logWithRequestID(requestID, "ERROR", "FIREWALL", "Failed to create client request: %v", err)
			sendError(w, "Failed to create client request: "+err.Error(), http.StatusInternalServerError)
			return
		}

		clientReq.Header.Set("Authorization", "Bearer "+device.AccessToken)
		clientReq.Header.Set("Content-Type", "application/json")
		clientReq.Header.Set("X-Request-ID", requestID)

		// Add timeout to prevent hanging
		httpClient := &http.Client{
			Timeout: 30 * time.Second,
		}

		// Log timing information
		startTime := time.Now()
		logWithRequestID(requestID, "INFO", "FIREWALL", "Sending HTTP request to client...")

		resp, err := httpClient.Do(clientReq)
		duration := time.Since(startTime)

		if err != nil {
			logWithRequestID(requestID, "ERROR", "FIREWALL", "Failed to reach client after %v: %v", duration, err)
			sendError(w, "Failed to reach client: "+err.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		logWithRequestID(requestID, "INFO", "FIREWALL", "Client responded in %v with status: %d", duration, resp.StatusCode)

		// Handle client error responses
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			logWithRequestID(requestID, "ERROR", "FIREWALL", "Client returned non-200 status: %d", resp.StatusCode)

			var clientError ErrorResponse
			if json.Unmarshal(body, &clientError) == nil && clientError.Status == "failed" {
				sendError(w, "Client error: "+clientError.Message, http.StatusBadGateway)
			} else {
				sendError(w, "Client error: "+string(body), http.StatusBadGateway)
			}
			return
		}

		// Read client response
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			logWithRequestID(requestID, "ERROR", "FIREWALL", "Failed to read response body: %v", err)
			sendError(w, "Failed to read client response: "+err.Error(), http.StatusBadGateway)
			return
		}

		logWithRequestID(requestID, "DEBUG", "FIREWALL", "Client response body: %s", string(body))

		// Parse client response
		var clientResp interface{}
		if err := json.Unmarshal(body, &clientResp); err != nil {
			logWithRequestID(requestID, "ERROR", "FIREWALL", "Failed to parse client response: %v", err)
			sendError(w, "Invalid client response: "+err.Error(), http.StatusBadGateway)
			return
		}

		// Process response based on OS
		processedResp := processFirewallUpdateResponse(clientResp, device.Os)

		// Send successful response
		sendGetSuccess(w, processedResp)
		logWithRequestID(requestID, "INFO", "FIREWALL", "✅ Firewall %s completed successfully for %s in %v", req.Action, req.Host, duration)
		logWithRequestID(requestID, "INFO", "FIREWALL", "=== FIREWALL REQUEST END ===")
	}
}

// Process firewall update request based on OS
func processFirewallUpdateRequest(req firewallUpdateRequest, osType string, requestID string) map[string]interface{} {
	clientPayload := map[string]interface{}{
		"action": req.Action,
	}

	if strings.ToLower(osType) == "windows" {
		logWithRequestID(requestID, "INFO", "FIREWALL", "🪟 Processing Windows firewall request")
		return processWindowsFirewallUpdateRequest(req, clientPayload, requestID)
	}

	// Default Linux behavior
	logWithRequestID(requestID, "INFO", "FIREWALL", "🐧 Processing Linux firewall request")
	return processLinuxFirewallUpdateRequest(req, clientPayload, requestID)
}

// Process Linux firewall update request
func processLinuxFirewallUpdateRequest(req firewallUpdateRequest, clientPayload map[string]interface{}, requestID string) map[string]interface{} {
	linuxFields := []string{}

	// For Linux, always send the rule field for backward compatibility
	if req.Rule != "" {
		clientPayload["rule"] = req.Rule
		linuxFields = append(linuxFields, fmt.Sprintf("rule=%s", req.Rule))
	} else {
		// Default rule if not provided
		clientPayload["rule"] = "accept"
		linuxFields = append(linuxFields, "rule=accept(default)")
	}

	if req.Protocol != "" {
		clientPayload["protocol"] = req.Protocol
		linuxFields = append(linuxFields, fmt.Sprintf("protocol=%s", req.Protocol))
	}
	if req.Port != "" {
		clientPayload["port"] = req.Port
		linuxFields = append(linuxFields, fmt.Sprintf("port=%s", req.Port))
	}
	if req.Source != "" {
		clientPayload["source"] = req.Source
		linuxFields = append(linuxFields, fmt.Sprintf("source=%s", req.Source))
	}
	if req.Destination != "" {
		clientPayload["destination"] = req.Destination
		linuxFields = append(linuxFields, fmt.Sprintf("destination=%s", req.Destination))
	}

	logWithRequestID(requestID, "DEBUG", "FIREWALL", "Linux fields: [%s]", strings.Join(linuxFields, ", "))
	return clientPayload
}

// Process Windows firewall update request
func processWindowsFirewallUpdateRequest(req firewallUpdateRequest, clientPayload map[string]interface{}, requestID string) map[string]interface{} {
	windowsFields := []string{}

	if req.Name != "" {
		clientPayload["name"] = req.Name
		windowsFields = append(windowsFields, fmt.Sprintf("name=%s", req.Name))
	}
	if req.DisplayName != "" {
		clientPayload["displayName"] = req.DisplayName
		windowsFields = append(windowsFields, fmt.Sprintf("displayName=%s", req.DisplayName))
	}
	if req.Direction != "" {
		clientPayload["direction"] = req.Direction
		windowsFields = append(windowsFields, fmt.Sprintf("direction=%s", req.Direction))
	}
	if req.ActionType != "" {
		clientPayload["actionType"] = req.ActionType
		windowsFields = append(windowsFields, fmt.Sprintf("actionType=%s", req.ActionType))
	}
	if req.Enabled != "" {
		clientPayload["enabled"] = req.Enabled
		windowsFields = append(windowsFields, fmt.Sprintf("enabled=%s", req.Enabled))
	}
	if req.Profile != "" {
		clientPayload["profile"] = req.Profile
		windowsFields = append(windowsFields, fmt.Sprintf("profile=%s", req.Profile))
	}
	if req.Protocol != "" {
		clientPayload["protocol"] = req.Protocol
		windowsFields = append(windowsFields, fmt.Sprintf("protocol=%s", req.Protocol))
	}
	if req.LocalPort != "" {
		clientPayload["localPort"] = req.LocalPort
		windowsFields = append(windowsFields, fmt.Sprintf("localPort=%s", req.LocalPort))
	}
	if req.RemotePort != "" {
		clientPayload["remotePort"] = req.RemotePort
		windowsFields = append(windowsFields, fmt.Sprintf("remotePort=%s", req.RemotePort))
	}
	if req.LocalAddress != "" {
		clientPayload["localAddress"] = req.LocalAddress
		windowsFields = append(windowsFields, fmt.Sprintf("localAddress=%s", req.LocalAddress))
	}
	if req.RemoteAddress != "" {
		clientPayload["remoteAddress"] = req.RemoteAddress
		windowsFields = append(windowsFields, fmt.Sprintf("remoteAddress=%s", req.RemoteAddress))
	}
	if req.Program != "" {
		clientPayload["program"] = req.Program
		windowsFields = append(windowsFields, fmt.Sprintf("program=%s", req.Program))
	}
	if req.Service != "" {
		clientPayload["service"] = req.Service
		windowsFields = append(windowsFields, fmt.Sprintf("service=%s", req.Service))
	}

	logWithRequestID(requestID, "DEBUG", "FIREWALL", "Windows fields: [%s]", strings.Join(windowsFields, ", "))
	return clientPayload
}

// Process firewall update response based on OS
func processFirewallUpdateResponse(resp interface{}, osType string) interface{} {
	if strings.ToLower(osType) == "windows" {
		return processWindowsFirewallUpdateResponse(resp)
	}

	// Default Linux behavior
	return resp
}

// Windows-specific firewall update response processing (placeholder for future differences)
func processWindowsFirewallUpdateResponse(resp interface{}) interface{} {
	// For now, return same format as Linux
	// Future: might need different response handling for Windows Firewall
	return resp
}
