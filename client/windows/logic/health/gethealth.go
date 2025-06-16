package health

import (
	"encoding/json"
	"net"
	"net/http"
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

// getActiveInterface returns the primary non-loopback, non-virtual, UP interface
func getActiveInterface() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagLoopback != 0 ||
			iface.Flags&net.FlagUp == 0 ||
			strings.Contains(iface.Name, "vEthernet") || // Windows virtual interfaces
			strings.Contains(iface.Name, "Loopback") {
			continue
		}
		return iface.Name, nil
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

func HandleHealthConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var (
		cpuPercent []float64
		vmStat     *mem.VirtualMemoryStat
		diskStat   *disk.UsageStat
		activeIface string
		netStat    *NetStats
		conns      []psnet.ConnectionStat
		err        error
	)

	// CPU
	cpuPercent, err = cpu.Percent(time.Second, false)
	if err != nil {
		sendHealthFailure(w, "Failed to get CPU stats", err)
		return
	}

	// RAM
	vmStat, err = mem.VirtualMemory()
	if err != nil {
		sendHealthFailure(w, "Failed to get memory stats", err)
		return
	}

	// Disk
	diskStat, err = disk.Usage("C:\\")
	if err != nil {
		sendHealthFailure(w, "Failed to get disk stats", err)
		return
	}

	// Network Interface
	activeIface, err = getActiveInterface()
	if err != nil {
		sendHealthFailure(w, "Failed to get active interface", err)
		return
	}
	if activeIface != "" {
		netIOs, _ := psnet.IOCounters(true)
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

	// Open Ports
	conns, err = psnet.Connections("inet")
	if err != nil {
		sendHealthFailure(w, "Failed to get open ports", err)
		return
	}
	openPorts := []OpenPort{}
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

	// Compose health stats
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

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "System health metrics retrieved",
		"data":    stats,
	})
}

func sendHealthFailure(w http.ResponseWriter, msg string, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "failed",
		"message": msg,
		"data":    err.Error(),
	})
}
