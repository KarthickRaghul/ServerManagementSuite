package optimisation

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"

	"backend/config"
	serverdb "backend/db/gen/server"
)

type restartServiceRequest struct {
	Host    string `json:"host"`
	Service string `json:"service"`
}

type responseJSON1 struct {
	Status string `json:"status"`
}

func PostRestartService(queries *serverdb.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(responseJSON1{Status: "failure"})
			return
		}

		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(responseJSON1{Status: "failure"})
			return
		}

		var req restartServiceRequest
		if err := json.Unmarshal(bodyBytes, &req); err != nil || req.Host == "" || req.Service == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(responseJSON1{Status: "failure"})
			return
		}

		device, err := queries.GetServerDeviceByIP(context.Background(), req.Host)
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(responseJSON1{Status: "failure"})
			return
		} else if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(responseJSON1{Status: "failure"})
			return
		}

		// ✅ Use config for client URL (reads from .env file)
		clientURL := config.GetClientURL(req.Host, "/client/restartservice")

		clientReq, err := http.NewRequest("POST", clientURL, bytes.NewReader(bodyBytes))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(responseJSON1{Status: "failure"})
			return
		}
		clientReq.Header.Set("Authorization", "Bearer "+device.AccessToken)
		clientReq.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(clientReq)
		if err != nil || resp.StatusCode != http.StatusOK {
			w.WriteHeader(http.StatusBadGateway)
			json.NewEncoder(w).Encode(responseJSON1{Status: "failure"})
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
