package routine

import (
	"time"
)

type HealthResponse struct {
	CPU       CPUInfo    `json:"cpu"`
	RAM       RAMInfo    `json:"ram"`
	Disk      DiskInfo   `json:"disk"`
	Net       NetInfo    `json:"net"`
	OpenPorts []PortInfo `json:"open_ports"`
}

type CPUInfo struct {
	UsagePercent float64 `json:"usage_percent"`
}

type RAMInfo struct {
	TotalMB      float64 `json:"total_mb"`
	UsedMB       float64 `json:"used_mb"`
	FreeMB       float64 `json:"free_mb"`
	UsagePercent float64 `json:"usage_percent"`
}

type DiskInfo struct {
	TotalMB      float64 `json:"total_mb"`
	UsedMB       float64 `json:"used_mb"`
	FreeMB       float64 `json:"free_mb"`
	UsagePercent float64 `json:"usage_percent"`
}

type NetInfo struct {
	Name        string  `json:"name"`
	BytesSentMB float64 `json:"bytes_sent_mb"`
	BytesRecvMB float64 `json:"bytes_recv_mb"`
}

type PortInfo struct {
	Protocol string `json:"protocol"`
	Port     int    `json:"port"`
	Process  string `json:"process"`
}

type AlertRule struct {
	CPUThreshold  float64
	RAMThreshold  float64
	DiskThreshold float64
	CheckInterval time.Duration
}
