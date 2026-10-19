package api

import (
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
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
		if i >= 10 {
			break
		}
		procInfoList = append(procInfoList, p)
	}
	return procInfoList
}

func GetTotalCPU() float64 {
	percentage, err := cpu.Percent(time.Second, false)
	if err != nil {
		panic("Failed to get CPU usage")
	}
	return percentage[0]
}
