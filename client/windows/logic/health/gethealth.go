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
		if strings.Contains(iface.Name, "docker") || strings.Contains(iface.Name, "veth") ||
			strings.Contains(iface.Name, "br-") || strings.Contains(iface.Name, "virbr") {
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

func HandleHealthConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// CPU
	cpuPercent, _ := cpu.Percent(time.Second, false)

	// RAM
	vmStat, _ := mem.VirtualMemory()

	// Disk
	var diskPath string
	if runtime.GOOS == "windows" {
		diskPath = "C:\\"
	} else {
		diskPath = "/"
	}
	diskStat, _ := disk.Usage(diskPath)

	// Network
	activeIface, _ := getActiveInterface()
	var netStat *NetStats
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

	// Open ports
	conns, _ := psnet.Connections("inet")
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

	// Compose result
	stats := HealthStats{
		CPU: CPUStats{
			UsagePercent: 0,
		},
		RAM: RAMStats{
			Total:        0,
			Used:         0,
			Free:         0,
			UsagePercent: 0,
		},
		Disk: DiskStats{
			Total:        0,
			Used:         0,
			Free:         0,
			UsagePercent: 0,
		},
		Net:       netStat,
		OpenPorts: openPorts,
	}

	if len(cpuPercent) > 0 {
		stats.CPU.UsagePercent = cpuPercent[0]
	}
	if vmStat != nil {
		stats.RAM.Total = bytesToMB(vmStat.Total)
		stats.RAM.Used = bytesToMB(vmStat.Used)
		stats.RAM.Free = bytesToMB(vmStat.Free)
		stats.RAM.UsagePercent = vmStat.UsedPercent
	}
	if diskStat != nil {
		stats.Disk.Total = bytesToMB(diskStat.Total)
		stats.Disk.Used = bytesToMB(diskStat.Used)
		stats.Disk.Free = bytesToMB(diskStat.Free)
		stats.Disk.UsagePercent = diskStat.UsedPercent
	}

	json.NewEncoder(w).Encode(stats)
}
