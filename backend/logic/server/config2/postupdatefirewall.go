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

	// ✅ Windows-specific fields
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

type responseJSON2 struct {
	Status  string `json:"status"`
	Error   string `json:"error,omitempty"`
	Details string `json:"details,omitempty"`
	Debug   string `json:"debug,omitempty"`   // ✅ Added debug field
	Payload string `json:"payload,omitempty"` // ✅ Added payload field for debugging
}

// ✅ Enhanced logging function with request ID
func logWithRequestID(requestID, level, component, message string, args ...interface{}) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	formattedMessage := fmt.Sprintf(message, args...)
	log.Printf("[%s] [%s] [%s] [%s] %s", timestamp, requestID, level, component, formattedMessage)
}

// ✅ Generate unique request ID
func generateRequestID() string {
	return fmt.Sprintf("fw-%d", time.Now().UnixNano())
}

func HandlePostUpdateFirewall(queries *serverdb.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// ✅ Generate unique request ID for tracing
		requestID := generateRequestID()

		// ✅ Enhanced request logging
		logWithRequestID(requestID, "INFO", "FIREWALL", "=== FIREWALL REQUEST START ===")
		logWithRequestID(requestID, "INFO", "FIREWALL", "Method: %s, URL: %s, RemoteAddr: %s", r.Method, r.URL.Path, r.RemoteAddr)
		logWithRequestID(requestID, "INFO", "FIREWALL", "User-Agent: %s", r.Header.Get("User-Agent"))
		logWithRequestID(requestID, "INFO", "FIREWALL", "Content-Type: %s", r.Header.Get("Content-Type"))

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Request-ID", requestID) // ✅ Add request ID to response headers

		if r.Method != http.MethodPost {
			logWithRequestID(requestID, "ERROR", "FIREWALL", "Method not allowed: %s", r.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(responseJSON2{
				Status: "failure",
				Error:  "Method not allowed",
				Debug:  fmt.Sprintf("Expected POST, got %s", r.Method),
			})
			return
		}

		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			logWithRequestID(requestID, "ERROR", "FIREWALL", "Failed to read request body: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(responseJSON2{
				Status: "failure",
				Error:  "Failed to read request body",
				Debug:  err.Error(),
			})
			return
		}

		// ✅ Log raw request body for debugging
		logWithRequestID(requestID, "DEBUG", "FIREWALL", "Raw request body: %s", string(bodyBytes))

		var req firewallUpdateRequest
		if err := json.Unmarshal(bodyBytes, &req); err != nil {
			logWithRequestID(requestID, "ERROR", "FIREWALL", "Failed to unmarshal request: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(responseJSON2{
				Status:  "failure",
				Error:   "Invalid request format",
				Debug:   fmt.Sprintf("JSON unmarshal error: %v", err),
				Payload: string(bodyBytes),
			})
			return
		}

		// ✅ Log parsed request structure
		logWithRequestID(requestID, "DEBUG", "FIREWALL", "Parsed request: %+v", req)

		// Basic validation
		if req.Host == "" || req.Action == "" {
			logWithRequestID(requestID, "ERROR", "FIREWALL", "Missing required fields - Host: '%s', Action: '%s'", req.Host, req.Action)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(responseJSON2{
				Status: "failure",
				Error:  "Host and action are required",
				Debug:  fmt.Sprintf("Host='%s', Action='%s'", req.Host, req.Action),
			})
			return
		}

		logWithRequestID(requestID, "INFO", "FIREWALL", "Processing firewall %s for host: %s", req.Action, req.Host)

		device, err := queries.GetServerDeviceByIP(context.Background(), req.Host)
		if err == sql.ErrNoRows {
			logWithRequestID(requestID, "ERROR", "FIREWALL", "Device not found: %s", req.Host)
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(responseJSON2{
				Status: "failure",
				Error:  "Device not registered",
				Debug:  fmt.Sprintf("No device found with IP: %s", req.Host),
			})
			return
		} else if err != nil {
			logWithRequestID(requestID, "ERROR", "FIREWALL", "Database error: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(responseJSON2{
				Status: "failure",
				Error:  "Database error",
				Debug:  err.Error(),
			})
			return
		}

		logWithRequestID(requestID, "DEBUG", "FIREWALL", "Found device: ID=%s, Tag=%s", device.ID, device.Tag)

		// ✅ Create client payload WITHOUT the host field
		clientPayload := map[string]interface{}{
			"action": req.Action,
		}

		// ✅ Enhanced firewall type detection with detailed logging
		isWindowsFirewall := req.Name != "" || req.DisplayName != "" || req.Direction != "" || req.ActionType != ""

		if isWindowsFirewall {
			// Windows firewall request
			logWithRequestID(requestID, "INFO", "FIREWALL", "🪟 Detected Windows firewall request")

			// ✅ Log each Windows field being processed
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
		} else {
			// ✅ Linux firewall request - use legacy format for compatibility
			logWithRequestID(requestID, "INFO", "FIREWALL", "🐧 Detected Linux firewall request")

			linuxFields := []string{}

			// ✅ For Linux, always send the rule field for backward compatibility
			if req.Rule != "" {
				clientPayload["rule"] = req.Rule
				linuxFields = append(linuxFields, fmt.Sprintf("rule=%s", req.Rule))
			} else {
				// ✅ Default rule if not provided
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
		}

		// Convert to JSON for client
		clientBody, err := json.Marshal(clientPayload)
		if err != nil {
			logWithRequestID(requestID, "ERROR", "FIREWALL", "Failed to marshal client payload: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(responseJSON2{
				Status: "failure",
				Error:  "Failed to prepare client request",
				Debug:  err.Error(),
			})
			return
		}

		logWithRequestID(requestID, "DEBUG", "FIREWALL", "Client payload: %s", string(clientBody))

		// Use config for client URL
		clientURL := config.GetClientURL(req.Host, "/client/config2/updatefirewall")
		logWithRequestID(requestID, "INFO", "FIREWALL", "Sending request to: %s", clientURL)

		clientReq, err := http.NewRequest("POST", clientURL, bytes.NewReader(clientBody))
		if err != nil {
			logWithRequestID(requestID, "ERROR", "FIREWALL", "Failed to create client request: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(responseJSON2{
				Status: "failure",
				Error:  "Failed to create client request",
				Debug:  err.Error(),
			})
			return
		}

		clientReq.Header.Set("Authorization", "Bearer "+device.AccessToken)
		clientReq.Header.Set("Content-Type", "application/json")
		clientReq.Header.Set("X-Request-ID", requestID) // ✅ Forward request ID

		// ✅ Log request headers
		logWithRequestID(requestID, "DEBUG", "FIREWALL", "Client request headers: %v", clientReq.Header)

		// Add timeout to prevent hanging
		httpClient := &http.Client{
			Timeout: 30 * time.Second,
		}

		// ✅ Log timing information
		startTime := time.Now()
		logWithRequestID(requestID, "INFO", "FIREWALL", "Sending HTTP request to client...")

		resp, err := httpClient.Do(clientReq)
		duration := time.Since(startTime)

		if err != nil {
			logWithRequestID(requestID, "ERROR", "FIREWALL", "Failed to reach client after %v: %v", duration, err)
			w.WriteHeader(http.StatusBadGateway)
			json.NewEncoder(w).Encode(responseJSON2{
				Status:  "failure",
				Error:   "Failed to reach client",
				Details: err.Error(),
				Debug:   fmt.Sprintf("Request duration: %v, URL: %s", duration, clientURL),
			})
			return
		}
		defer resp.Body.Close()

		logWithRequestID(requestID, "INFO", "FIREWALL", "Client responded in %v with status: %d", duration, resp.StatusCode)

		// ✅ Log response headers
		logWithRequestID(requestID, "DEBUG", "FIREWALL", "Client response headers: %v", resp.Header)

		// Read response body for debugging
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			logWithRequestID(requestID, "ERROR", "FIREWALL", "Failed to read response body: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(responseJSON2{
				Status: "failure",
				Error:  "Failed to read client response",
				Debug:  err.Error(),
			})
			return
		}

		logWithRequestID(requestID, "DEBUG", "FIREWALL", "Client response body: %s", string(body))

		if resp.StatusCode != http.StatusOK {
			logWithRequestID(requestID, "ERROR", "FIREWALL", "Client returned non-200 status: %d", resp.StatusCode)

			// ✅ Try to parse client error for better debugging
			var clientError map[string]interface{}
			if json.Unmarshal(body, &clientError) == nil {
				logWithRequestID(requestID, "DEBUG", "FIREWALL", "Parsed client error: %+v", clientError)
			}

			w.WriteHeader(http.StatusBadGateway)
			json.NewEncoder(w).Encode(responseJSON2{
				Status:  "failure",
				Error:   fmt.Sprintf("Client returned status %d", resp.StatusCode),
				Details: string(body),
				Debug:   fmt.Sprintf("Request ID: %s, Duration: %v, Client URL: %s", requestID, duration, clientURL),
			})
			return
		}

		// Success - forward client response
		w.WriteHeader(resp.StatusCode)
		w.Write(body)
		logWithRequestID(requestID, "INFO", "FIREWALL", "✅ Firewall %s completed successfully for %s in %v", req.Action, req.Host, duration)
		logWithRequestID(requestID, "INFO", "FIREWALL", "=== FIREWALL REQUEST END ===")
	}
}
