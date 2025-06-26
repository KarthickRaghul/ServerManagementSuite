package health

import (
	"encoding/json"
	"net"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	psnet "github.com/shirou/gopsutil/v3/net"
	psproc "github.com/shirou/gopsutil/v3/process"
)

type CPUStats struct {
	UsagePercent float64 `json:"usage_percent"`
}

type RAMStats struct {
	Total        float64 `json:"total_mb"`
	Used         float64 `json:"used_mb"`
	Free         float64 `json:"free_mb"`
	UsagePercent float64 `json:"usage_percent"`
}

type DiskStats struct {
	Total        float64 `json:"total_mb"`
	Used         float64 `json:"used_mb"`
	Free         float64 `json:"free_mb"`
	UsagePercent float64 `json:"usage_percent"`
}

type NetStats struct {
	Name      string  `json:"name"`
	BytesSent float64 `json:"bytes_sent_mb"`
	BytesRecv float64 `json:"bytes_recv_mb"`
}

type OpenPort struct {
	Protocol string `json:"protocol"`
	Port     uint32 `json:"port"`
	Process  string `json:"process"`
}

type HealthStats struct {
	CPU       CPUStats   `json:"cpu"`
	RAM       RAMStats   `json:"ram"`
	Disk      DiskStats  `json:"disk"`
	Net       *NetStats  `json:"net"`
	OpenPorts []OpenPort `json:"open_ports"`
}

type ErrorResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// sendGetSuccess sends successful GET response with data
func sendGetSuccess(w http.ResponseWriter, data interface{}) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)
}

// sendError sends standardized error response
func sendError(w http.ResponseWriter, message string, statusCode int) {
	w.WriteHeader(statusCode)
	errorResp := ErrorResponse{
		Status:  "failed",
		Message: message,
	}
	json.NewEncoder(w).Encode(errorResp)
}

func HandleHealthConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendError(w, "Only GET method allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	// CPU
	cpuPercent, err := cpu.Percent(time.Second, false)
	if err != nil || len(cpuPercent) == 0 {
		sendError(w, "Failed to get CPU stats", http.StatusInternalServerError)
		return
	}

	// RAM
	vmStat, err := mem.VirtualMemory()
	if err != nil {
		sendError(w, "Failed to get RAM stats", http.StatusInternalServerError)
		return
	}

	// Disk
	var diskPath string
	if runtime.GOOS == "windows" {
		diskPath = "C:\\"
	} else {
		diskPath = "/"
	}
	diskStat, err := disk.Usage(diskPath)
	if err != nil {
		sendError(w, "Failed to get disk stats", http.StatusInternalServerError)
		return
	}

	// Network
	activeIface, _ := getActiveInterface()
	var netStat *NetStats
	if activeIface != "" {
		netIOs, err := psnet.IOCounters(true)
		if err == nil {
			for _, iface := range netIOs {
				if iface.Name == activeIface {
					netStat = &NetStats{
						Name:      iface.Name,
						BytesSent: bytesToMB(iface.BytesSent),
						BytesRecv: bytesToMB(iface.BytesRecv),
					}
					break
				}
			}
		}
	}

	// Open ports
	conns, err := psnet.Connections("inet")
	openPorts := []OpenPort{}
	if err == nil {
		for _, conn := range conns {
			if conn.Status == "LISTEN" && conn.Laddr.Port != 0 {
				protocol := "tcp"
				if conn.Type == 2 {
					protocol = "udp"
				}
				processName := ""
				if conn.Pid > 0 {
					processName = getProcessName(conn.Pid)
				}
				openPorts = append(openPorts, OpenPort{
					Protocol: protocol,
					Port:     conn.Laddr.Port,
					Process:  processName,
				})
			}
		}
	}

	stats := HealthStats{
		CPU: CPUStats{
			UsagePercent: cpuPercent[0],
		},
		RAM: RAMStats{
			Total:        bytesToMB(vmStat.Total),
			Used:         bytesToMB(vmStat.Used),
			Free:         bytesToMB(vmStat.Free),
			UsagePercent: vmStat.UsedPercent,
		},
		Disk: DiskStats{
			Total:        bytesToMB(diskStat.Total),
			Used:         bytesToMB(diskStat.Used),
			Free:         bytesToMB(diskStat.Free),
			UsagePercent: diskStat.UsedPercent,
		},
		Net:       netStat,
		OpenPorts: openPorts,
	}

	sendGetSuccess(w, stats)
}

// getActiveInterface returns the primary UP, non-loopback, non-virtual interface
func getActiveInterface() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	var candidates []net.Interface
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		if strings.Contains(strings.ToLower(iface.Name), "docker") ||
			strings.Contains(strings.ToLower(iface.Name), "veth") ||
			strings.Contains(strings.ToLower(iface.Name), "br-") ||
			strings.Contains(strings.ToLower(iface.Name), "virbr") {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil || len(addrs) == 0 {
			continue
		}
		// Windows common names
		if strings.Contains(strings.ToLower(iface.Name), "ethernet") ||
			strings.Contains(strings.ToLower(iface.Name), "wi-fi") {
			return iface.Name, nil
		}
		// Linux/macOS: prioritize wireless/en
		if strings.HasPrefix(iface.Name, "wl") || strings.HasPrefix(iface.Name, "en") {
			return iface.Name, nil
		}
		candidates = append(candidates, iface)
	}
	if len(candidates) > 0 {
		return candidates[0].Name, nil
	}
	return "", nil
}

func bytesToMB(b uint64) float64 {
	return float64(b) / (1024 * 1024)
}

func getProcessName(pid int32) string {
	proc, err := psproc.NewProcess(pid)
	if err != nil {
		return ""
	}
	name, err := proc.Name()
	if err != nil {
		return ""
	}
	return name
}
