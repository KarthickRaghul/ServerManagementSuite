package config2

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	serverdb "backend/db/gen/server"
)

type hostExtract6 struct {
	Host string `json:"host"`
}

func HandleGetNetworkBasics(queries *serverdb.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req hostExtract6
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Host == "" {
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

		clientURL := fmt.Sprintf("http://%s/client/config2/network", req.Host)
		clientReq, err := http.NewRequest("GET", clientURL, nil)
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

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
	}
}
