package api

import (
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

type ResourceType int

const (
	CPU ResourceType = iota
	RAM
)

func UpdateMonitor() {
	findProcesses()
}

func GetHeadProcesses(n int, rt ResourceType) []ProcInfo {
	var procInfoList []ProcInfo

	for i, p := range sortProcesses(processes, rt) {
		if i >= n {
			break
		}
		procInfoList = append(procInfoList, p)
	}
	return procInfoList
}

func GetUsageCPU() float64 {
	percentage, err := cpu.Percent(time.Second, false)
	if err != nil {
		panic("Failed to get CPU usage")
	}
	return percentage[0]
}

func GetUsageRAM() float64 {
	info, err := mem.VirtualMemory()
	if err != nil {
		panic("Failed to get total RAM")
	}
	return info.UsedPercent
}

func GetTotalRAM() float64 {
	info, err := mem.VirtualMemory()
	if err != nil {
		panic("Failed to get RAM usage")
	}
	return float64(info.Available)
}
