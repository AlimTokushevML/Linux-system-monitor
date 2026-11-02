package api

import (
	"github.com/shirou/gopsutil/v3/process"
)

type ProcInfo struct {
	PID         int
	Name        string
	Consumption map[ResourceType]float64
	Net         NetUsage
}

func findProcesses() ([]ProcInfo, error) {
	procs, err := process.Processes()
	if err != nil {
		return nil, err
	}
	return collectProcesses(procs), nil
}

func collectProcesses(procs []*process.Process) []ProcInfo {
	result := make([]ProcInfo, 0, len(procs))
	for _, p := range procs {
		var newProcInfo ProcInfo
		var err error

		newProcInfo.PID = int(p.Pid)
		newProcInfo.Consumption = make(map[ResourceType]float64)

		newProcInfo.Name, err = p.Name()
		if err != nil {
			newProcInfo.Name = "Failed to retrieve the process"
		}
		newProcInfo.Consumption[CPU], err = p.CPUPercent()
		if err != nil {
			newProcInfo.Consumption[CPU] = 0
		}
		memInfo, err := p.MemoryInfo()
		if err != nil {
			newProcInfo.Consumption[RAM] = 0
		} else {
			newProcInfo.Consumption[RAM] = float64(memInfo.RSS)
		}

		result = append(result, newProcInfo)
	}
	return result
}
