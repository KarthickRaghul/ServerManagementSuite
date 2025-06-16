package config2

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	serverdb "backend/db/gen/server"
)

type firewallUpdateRequest struct {
	Host        string `json:"host"`
	Action      string `json:"action"`      // add/delete
	Rule        string `json:"rule"`        // accept/drop/reject
	Protocol    string `json:"protocol"`    // tcp/udp
	Port        string `json:"port"`        // e.g., "80"
	Source      string `json:"source"`      // optional
	Destination string `json:"destination"` // optional
}

type responseJSON2 struct {
	Status string `json:"status"`
}

func HandlePostUpdateFirewall(queries *serverdb.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(responseJSON2{Status: "failure"})
			return
		}

		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(responseJSON2{Status: "failure"})
			return
		}

		var req firewallUpdateRequest
		if err := json.Unmarshal(bodyBytes, &req); err != nil ||
			req.Host == "" || req.Action == "" || req.Rule == "" || req.Protocol == "" || req.Port == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(responseJSON2{Status: "failure"})
			return
		}

		device, err := queries.GetServerDeviceByIP(context.Background(), req.Host)
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(responseJSON2{Status: "failure"})
			return
		} else if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(responseJSON2{Status: "failure"})
			return
		}

		clientURL := fmt.Sprintf("http://%s/client/config2/updatefirewall", req.Host)
		clientReq, err := http.NewRequest("POST", clientURL, bytes.NewReader(bodyBytes))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(responseJSON2{Status: "failure"})
			return
		}
		clientReq.Header.Set("Authorization", "Bearer "+device.AccessToken)
		clientReq.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(clientReq)
		if err != nil || resp.StatusCode != http.StatusOK {
			w.WriteHeader(http.StatusBadGateway)
			json.NewEncoder(w).Encode(responseJSON2{Status: "failure"})
			return
		}
		defer resp.Body.Close()

		w.WriteHeader(resp.StatusCode)
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, "Failed to read response from client", http.StatusInternalServerError)
			return
		}
		w.Write(body)
	}
}
