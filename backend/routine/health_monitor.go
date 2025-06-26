package routine

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/kishore-001/ServerManagementSuite/backend/config"
	generaldb "github.com/kishore-001/ServerManagementSuite/backend/db/gen/general"
	serverdb "github.com/kishore-001/ServerManagementSuite/backend/db/gen/server"
)

// ✅ Updated type definitions with float64 for better JSON compatibility
type AlertRule struct {
	CPUThreshold  float64
	RAMThreshold  float64
	DiskThreshold float64
	CheckInterval time.Duration
}

type OpenPort struct {
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
	Process  string `json:"process"`
}

type CPUInfo struct {
	UsagePercent float64 `json:"usage_percent"`
	Cores        int     `json:"cores"`
}

// ✅ Changed from int64 to float64 to handle decimal numbers
type RAMInfo struct {
	UsagePercent float64 `json:"usage_percent"`
	TotalMB      float64 `json:"total_mb"` // ✅ Changed from int64 to float64
	UsedMB       float64 `json:"used_mb"`  // ✅ Changed from int64 to float64
}

// ✅ Changed from int64 to float64 to handle decimal numbers
type DiskInfo struct {
	UsagePercent float64 `json:"usage_percent"`
	TotalGB      float64 `json:"total_gb"` // ✅ Changed from int64 to float64
	UsedGB       float64 `json:"used_gb"`  // ✅ Changed from int64 to float64
}

// ✅ Changed from int64 to float64 to handle decimal numbers
type NetworkInfo struct {
	BytesReceived float64 `json:"bytes_received"` // ✅ Changed from int64 to float64
	BytesSent     float64 `json:"bytes_sent"`     // ✅ Changed from int64 to float64
}

type HealthResponse struct {
	CPU       CPUInfo     `json:"cpu"`
	RAM       RAMInfo     `json:"ram"`
	Disk      DiskInfo    `json:"disk"`
	Network   NetworkInfo `json:"network"`
	OpenPorts []OpenPort  `json:"open_ports"`
	Uptime    string      `json:"uptime"`
	Status    string      `json:"status"`
}

type HealthMonitor struct {
	queries   *serverdb.Queries
	rules     AlertRule
	client    *http.Client
	stopChan  chan bool
	isRunning bool

	// Alert suppression tracking
	lastAlerts    map[string]map[string]time.Time // host -> alert_type -> last_sent_time
	alertCounts   map[string]map[string]int       // host -> alert_type -> count
	suppressionMu sync.RWMutex

	// Suppression configuration
	suppressionRules SuppressionConfig

	// Email service
	emailService *EmailService
}

type SuppressionConfig struct {
	// Connectivity alerts
	ConnectivitySuppressionDuration time.Duration // Don't repeat connectivity alerts for this duration
	ConnectivityMaxBurst            int           // Max connectivity alerts before longer suppression
	ConnectivityBurstWindow         time.Duration // Time window for burst detection

	// Metric alerts
	MetricSuppressionDuration time.Duration // Don't repeat same metric alerts
	MetricEscalationThreshold int           // Escalate to critical after this many warnings

	// General
	CleanupInterval time.Duration // Clean old suppression data
}

func NewHealthMonitor(queries *serverdb.Queries, generalQueries *generaldb.Queries) *HealthMonitor {
	return &HealthMonitor{
		queries: queries,
		rules: AlertRule{
			CPUThreshold:  80.0,
			RAMThreshold:  85.0,
			DiskThreshold: 90.0,
			CheckInterval: 30 * time.Second,
		},
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		stopChan:    make(chan bool),
		lastAlerts:  make(map[string]map[string]time.Time),
		alertCounts: make(map[string]map[string]int),

		// Configure suppression rules
		suppressionRules: SuppressionConfig{
			ConnectivitySuppressionDuration: 5 * time.Minute,  // Don't repeat connectivity alerts for 5 minutes
			ConnectivityMaxBurst:            3,                // Max 3 alerts before longer suppression
			ConnectivityBurstWindow:         15 * time.Minute, // 15-minute window for burst detection
			MetricSuppressionDuration:       10 * time.Minute, // Don't repeat metric alerts for 10 minutes
			MetricEscalationThreshold:       5,                // Escalate after 5 consecutive warnings
			CleanupInterval:                 1 * time.Hour,    // Clean old data every hour
		},

		// Initialize email service
		emailService: NewEmailService(generalQueries),
	}
}

func (hm *HealthMonitor) Start() {
	if hm.isRunning {
		return
	}

	hm.isRunning = true
	log.Println("🔍 Health Monitor started with alert suppression and email notifications")

	go hm.monitorLoop()
	go hm.cleanupLoop() // Clean old suppression data
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

func (hm *HealthMonitor) cleanupLoop() {
	ticker := time.NewTicker(hm.suppressionRules.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			hm.cleanupOldSuppressionData()
		case <-hm.stopChan:
			return
		}
	}
}

func (hm *HealthMonitor) checkAllDevices() {
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
	healthData, err := hm.getHealthData(host, accessToken)
	if err != nil {
		// Handle connectivity alert with suppression
		hm.handleConnectivityAlert(host, err)
		return
	}

	// Device is reachable - reset connectivity alert count
	hm.resetAlertCount(host, "connectivity")

	// Check for health-based alerts
	hm.evaluateHealthRules(host, healthData)
}

// ✅ Enhanced getHealthData with better error handling and debugging
func (hm *HealthMonitor) getHealthData(host, accessToken string) (*HealthResponse, error) {
	url := config.GetClientURL(host, "/client/health")

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := hm.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("network error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("client returned status %d", resp.StatusCode)
	}

	// ✅ Read response body for debugging
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	var healthData HealthResponse
	if err := json.Unmarshal(bodyBytes, &healthData); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %v", err)
	}

	return &healthData, nil
}

func (hm *HealthMonitor) handleConnectivityAlert(host string, err error) {
	alertType := "connectivity"

	// Check if we should suppress this alert
	if hm.shouldSuppressAlert(host, alertType) {
		// Increment count but don't send alert
		hm.incrementAlertCount(host, alertType)
		log.Printf("🔇 Suppressed connectivity alert for %s (count: %d)", host, hm.getAlertCount(host, alertType))
		return
	}

	// Determine severity based on alert count
	count := hm.getAlertCount(host, alertType)
	severity := "warning"

	if count >= hm.suppressionRules.ConnectivityMaxBurst {
		severity = "critical"
	}

	content := fmt.Sprintf("Device unreachable (attempt %d): %v", count+1, err)

	// Create the alert
	hm.createSuppressedAlert(host, alertType, severity, content)
}

func (hm *HealthMonitor) evaluateHealthRules(host string, health *HealthResponse) {
	// Check CPU usage with suppression
	if health.CPU.UsagePercent > hm.rules.CPUThreshold {
		alertType := "cpu_high"
		if !hm.shouldSuppressAlert(host, alertType) {
			content := fmt.Sprintf("High CPU usage: %.2f%% (threshold: %.2f%%)",
				health.CPU.UsagePercent, hm.rules.CPUThreshold)

			severity := hm.determineSeverity(host, alertType, "warning")
			hm.createSuppressedAlert(host, alertType, severity, content)
		}
	} else {
		// Reset CPU alert count when back to normal
		hm.resetAlertCount(host, "cpu_high")
	}

	// Check RAM usage with suppression
	if health.RAM.UsagePercent > hm.rules.RAMThreshold {
		alertType := "ram_high"
		if !hm.shouldSuppressAlert(host, alertType) {
			content := fmt.Sprintf("High RAM usage: %.2f%% (threshold: %.2f%%)",
				health.RAM.UsagePercent, hm.rules.RAMThreshold)

			severity := hm.determineSeverity(host, alertType, "warning")
			hm.createSuppressedAlert(host, alertType, severity, content)
		}
	} else {
		hm.resetAlertCount(host, "ram_high")
	}

	// Check Disk usage with suppression
	if health.Disk.UsagePercent > hm.rules.DiskThreshold {
		alertType := "disk_high"
		if !hm.shouldSuppressAlert(host, alertType) {
			content := fmt.Sprintf("High Disk usage: %.2f%% (threshold: %.2f%%)",
				health.Disk.UsagePercent, hm.rules.DiskThreshold)

			severity := hm.determineSeverity(host, alertType, "critical")
			hm.createSuppressedAlert(host, alertType, severity, content)
		}
	} else {
		hm.resetAlertCount(host, "disk_high")
	}

	// Check suspicious ports (less frequent alerts)
	hm.checkSuspiciousPorts(host, health.OpenPorts)
}

func (hm *HealthMonitor) checkSuspiciousPorts(host string, openPorts []OpenPort) {
	suspiciousPorts := []int{22, 23, 3389}

	for _, port := range openPorts {
		for _, suspicious := range suspiciousPorts {
			if port.Port == suspicious && port.Protocol == "tcp" {
				alertType := fmt.Sprintf("suspicious_port_%d", port.Port)

				// Only alert once per day for suspicious ports
				if !hm.shouldSuppressAlertWithDuration(host, alertType, 24*time.Hour) {
					content := fmt.Sprintf("Suspicious port open: %d (%s) - Process: %s",
						port.Port, port.Protocol, port.Process)
					hm.createSuppressedAlert(host, alertType, "info", content)
				}
			}
		}
	}
}

// Alert suppression helper methods
func (hm *HealthMonitor) shouldSuppressAlert(host, alertType string) bool {
	return hm.shouldSuppressAlertWithDuration(host, alertType, hm.getSuppressionDuration(alertType))
}

func (hm *HealthMonitor) shouldSuppressAlertWithDuration(host, alertType string, duration time.Duration) bool {
	hm.suppressionMu.RLock()
	defer hm.suppressionMu.RUnlock()

	if hostAlerts, exists := hm.lastAlerts[host]; exists {
		if lastTime, exists := hostAlerts[alertType]; exists {
			return time.Since(lastTime) < duration
		}
	}
	return false
}

func (hm *HealthMonitor) getSuppressionDuration(alertType string) time.Duration {
	switch alertType {
	case "connectivity":
		return hm.suppressionRules.ConnectivitySuppressionDuration
	default:
		return hm.suppressionRules.MetricSuppressionDuration
	}
}

func (hm *HealthMonitor) determineSeverity(host, alertType, defaultSeverity string) string {
	count := hm.getAlertCount(host, alertType)

	// Escalate severity after multiple consecutive alerts
	if count >= hm.suppressionRules.MetricEscalationThreshold {
		if defaultSeverity == "warning" {
			return "critical"
		}
	}

	return defaultSeverity
}

// ✅ Updated createSuppressedAlert with email functionality
func (hm *HealthMonitor) createSuppressedAlert(host, alertType, severity, content string) {
	// Update suppression tracking
	hm.updateSuppressionTracking(host, alertType)

	// Create the alert in database
	_, err := hm.queries.CreateAlert(context.Background(), serverdb.CreateAlertParams{
		Host:     host,
		Severity: severity,
		Content:  content,
	})

	if err != nil {
		log.Printf("❌ Failed to create alert for %s: %v", host, err)
		return
	}

	count := hm.getAlertCount(host, alertType)
	log.Printf("🚨 Alert created for %s [%s] (count: %d): %s", host, severity, count, content)

	// ✅ Send email notification asynchronously
	go func() {
		if err := hm.emailService.SendAlertEmail(host, severity, content); err != nil {
			log.Printf("❌ Failed to send email for alert: %v", err)
		}
	}()
}

func (hm *HealthMonitor) updateSuppressionTracking(host, alertType string) {
	hm.suppressionMu.Lock()
	defer hm.suppressionMu.Unlock()

	// Initialize maps if needed
	if hm.lastAlerts[host] == nil {
		hm.lastAlerts[host] = make(map[string]time.Time)
	}
	if hm.alertCounts[host] == nil {
		hm.alertCounts[host] = make(map[string]int)
	}

	// Update last alert time and increment count
	hm.lastAlerts[host][alertType] = time.Now()
	hm.alertCounts[host][alertType]++
}

func (hm *HealthMonitor) incrementAlertCount(host, alertType string) {
	hm.suppressionMu.Lock()
	defer hm.suppressionMu.Unlock()

	if hm.alertCounts[host] == nil {
		hm.alertCounts[host] = make(map[string]int)
	}
	hm.alertCounts[host][alertType]++
}

func (hm *HealthMonitor) getAlertCount(host, alertType string) int {
	hm.suppressionMu.RLock()
	defer hm.suppressionMu.RUnlock()

	if hostCounts, exists := hm.alertCounts[host]; exists {
		return hostCounts[alertType]
	}
	return 0
}

func (hm *HealthMonitor) resetAlertCount(host, alertType string) {
	hm.suppressionMu.Lock()
	defer hm.suppressionMu.Unlock()

	if hm.alertCounts[host] != nil {
		delete(hm.alertCounts[host], alertType)
	}
}

func (hm *HealthMonitor) cleanupOldSuppressionData() {
	hm.suppressionMu.Lock()
	defer hm.suppressionMu.Unlock()

	cutoff := time.Now().Add(-24 * time.Hour) // Remove data older than 24 hours

	for host, hostAlerts := range hm.lastAlerts {
		for alertType, lastTime := range hostAlerts {
			if lastTime.Before(cutoff) {
				delete(hostAlerts, alertType)
			}
		}

		// Clean empty host entries
		if len(hostAlerts) == 0 {
			delete(hm.lastAlerts, host)
		}
	}

	// Also clean alert counts for consistency
	for host := range hm.alertCounts {
		if _, exists := hm.lastAlerts[host]; !exists {
			delete(hm.alertCounts, host)
		}
	}

	log.Println("🧹 Cleaned up old suppression data")
}

// ✅ Enhanced HandleEmail function (now calls email service)
func (hm *HealthMonitor) HandleEmail(host, severity, content string) {
	if err := hm.emailService.SendAlertEmail(host, severity, content); err != nil {
		log.Printf("❌ Failed to send email for %s [%s]: %v", host, severity, err)
	} else {
		log.Printf("📧 Email sent successfully for %s [%s]: %s", host, severity, content)
	}
}
