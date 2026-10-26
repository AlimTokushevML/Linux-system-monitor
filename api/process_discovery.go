package api

import (
	"cmp"
	"log"
	"slices"

	"github.com/shirou/gopsutil/v3/process"
)

var processes []ProcInfo

type ProcInfo struct {
	PID         int
	Name        string
	Consumption map[ResourceType]float64
}

func findProcesses() {
	procs, err := process.Processes()
	processes = processes[:0]
	if err != nil {
		log.Fatal("Failed to retrieve the processes")
	}
	storeProcesses(procs)
}

func storeProcesses(procs []*process.Process) {
	for _, p := range procs {
		var newProcInfo ProcInfo
		var err error

		newProcInfo.PID = int(p.Pid)
		newProcInfo.Consumption = make(map[ResourceType]float64) // ← fix

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
		}
		newProcInfo.Consumption[RAM] = float64(memInfo.RSS)

		processes = append(processes, newProcInfo)
	}
}

func sortProcesses(procs []ProcInfo, rt ResourceType) []ProcInfo {
	slices.SortFunc(procs, func(a, b ProcInfo) int {
		return cmp.Compare(b.Consumption[rt], a.Consumption[rt])
	})
	return procs
}
