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

type hostExtract2 struct {
	Host string `json:"host"`
}

func HandlePostInterface1(queries *serverdb.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}

		var req hostExtract2
		if err := json.Unmarshal(bodyBytes, &req); err != nil || req.Host == "" {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		device, err := queries.GetServerDeviceByIP(context.Background(), req.Host)
		if err == sql.ErrNoRows {
			http.Error(w, "Device not registered", http.StatusNotFound)
			return
		} else if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		clientURL := fmt.Sprintf("http://%s/client/config2/restartinterface", req.Host)
		clientReq, err := http.NewRequest("POST", clientURL, bytes.NewReader(bodyBytes))
		if err != nil {
			http.Error(w, "Failed to create request to client", http.StatusInternalServerError)
			return
		}
		clientReq.Header.Set("Authorization", "Bearer "+device.AccessToken)
		clientReq.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(clientReq)
		if err != nil {
			http.Error(w, "Failed to reach client", http.StatusBadGateway)
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
