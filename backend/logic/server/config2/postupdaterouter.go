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

type routerUpdateRequest struct {
	Host        string `json:"host"`
	Action      string `json:"action"`      // "add" or "delete"
	Destination string `json:"destination"` // e.g., "192.168.2.0/24"
	Gateway     string `json:"gateway"`     // e.g., "192.168.1.1"
	Interface   string `json:"interface"`   // optional
	Metric      string `json:"metric"`      // optional
}

type responseJSON3 struct {
	Status string `json:"status"`
}

func HandlePostUpdateRouter(queries *serverdb.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(responseJSON3{Status: "failure1"})
			return
		}

		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(responseJSON3{Status: "failure2"})
			return
		}

		var req routerUpdateRequest
		if err := json.Unmarshal(bodyBytes, &req); err != nil ||
			req.Host == "" || req.Action == "" || req.Destination == "" || req.Gateway == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(responseJSON3{Status: "failure3"})
			return
		}

		device, err := queries.GetServerDeviceByIP(context.Background(), req.Host)
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(responseJSON3{Status: "failure4"})
			return
		} else if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(responseJSON3{Status: "failure5"})
			return
		}

		clientURL := fmt.Sprintf("http://%s/client/config2/updateroute", req.Host)
		clientReq, err := http.NewRequest("POST", clientURL, bytes.NewReader(bodyBytes))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(responseJSON3{Status: "failure6"})
			return
		}
		clientReq.Header.Set("Authorization", "Bearer "+device.AccessToken)
		clientReq.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(clientReq)
		if err != nil || resp.StatusCode != http.StatusOK {
			w.WriteHeader(http.StatusBadGateway)
			json.NewEncoder(w).Encode(responseJSON3{Status: "failure7"})
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
