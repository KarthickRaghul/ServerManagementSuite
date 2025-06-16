package routine

import (
	serverdb "backend/db/gen/server"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type HealthMonitor struct {
	queries   *serverdb.Queries
	rules     AlertRule
	client    *http.Client
	stopChan  chan bool
	isRunning bool
}

func NewHealthMonitor(queries *serverdb.Queries) *HealthMonitor {
	return &HealthMonitor{
		queries: queries,
		rules: AlertRule{
			CPUThreshold:  80.0,             // 80% CPU usage
			RAMThreshold:  85.0,             // 85% RAM usage
			DiskThreshold: 90.0,             // 90% Disk usage
			CheckInterval: 30 * time.Second, // Check every 30 seconds
		},
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		stopChan: make(chan bool),
	}
}

func (hm *HealthMonitor) Start() {
	if hm.isRunning {
		return
	}

	hm.isRunning = true
	log.Println("🔍 Health Monitor started")

	go hm.monitorLoop()
}

func (hm *HealthMonitor) Stop() {
	if !hm.isRunning {
		return
	}

	hm.stopChan <- true
	hm.isRunning = false
	log.Println("⏹️ Health Monitor stopped")
}

func (hm *HealthMonitor) monitorLoop() {
	ticker := time.NewTicker(hm.rules.CheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			hm.checkAllDevices()
		case <-hm.stopChan:
			return
		}
	}
}

func (hm *HealthMonitor) checkAllDevices() {
	// Get all registered devices
	devices, err := hm.queries.GetAllServerDevices(context.Background())
	if err != nil {
		log.Printf("❌ Failed to get devices: %v", err)
		return
	}

	for _, device := range devices {
		go hm.checkDeviceHealth(device.Ip, device.AccessToken)
	}
}

func (hm *HealthMonitor) checkDeviceHealth(host, accessToken string) {
	// Get health data from client
	healthData, err := hm.getHealthData(host, accessToken)
	if err != nil {
		// Create connectivity alert
		hm.createAlert(host, "critical", fmt.Sprintf("Failed to reach device: %v", err))
		return
	}

	// Check for alerts based on rules
	hm.evaluateHealthRules(host, healthData)
}

func (hm *HealthMonitor) getHealthData(host, accessToken string) (*HealthResponse, error) {
	url := fmt.Sprintf("http://%s/client/health", host)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := hm.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("client returned status %d", resp.StatusCode)
	}

	var healthData HealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&healthData); err != nil {
		return nil, err
	}

	return &healthData, nil
}

func (hm *HealthMonitor) evaluateHealthRules(host string, health *HealthResponse) {
	// Check CPU usage
	if health.CPU.UsagePercent > hm.rules.CPUThreshold {
		content := fmt.Sprintf("High CPU usage: %.2f%% (threshold: %.2f%%)",
			health.CPU.UsagePercent, hm.rules.CPUThreshold)
		hm.createAlert(host, "warning", content)
	}

	// Check RAM usage
	if health.RAM.UsagePercent > hm.rules.RAMThreshold {
		content := fmt.Sprintf("High RAM usage: %.2f%% (threshold: %.2f%%)",
			health.RAM.UsagePercent, hm.rules.RAMThreshold)
		hm.createAlert(host, "warning", content)
	}

	// Check Disk usage
	if health.Disk.UsagePercent > hm.rules.DiskThreshold {
		content := fmt.Sprintf("High Disk usage: %.2f%% (threshold: %.2f%%)",
			health.Disk.UsagePercent, hm.rules.DiskThreshold)
		hm.createAlert(host, "critical", content)
	}

	// Check for suspicious open ports (example rule)
	suspiciousPorts := []int{22, 23, 3389} // SSH, Telnet, RDP
	for _, port := range health.OpenPorts {
		for _, suspicious := range suspiciousPorts {
			if port.Port == suspicious && port.Protocol == "tcp" {
				content := fmt.Sprintf("Suspicious port open: %d (%s) - Process: %s",
					port.Port, port.Protocol, port.Process)
				hm.createAlert(host, "info", content)
			}
		}
	}
}

func (hm *HealthMonitor) createAlert(host, severity, content string) {
	_, err := hm.queries.CreateAlert(context.Background(), serverdb.CreateAlertParams{
		Host:     host,
		Severity: severity,
		Content:  content,
	})

	if err != nil {
		log.Printf("❌ Failed to create alert for %s: %v", host, err)
	} else {
		log.Printf("🚨 Alert created for %s [%s]: %s", host, severity, content)
	}
}
