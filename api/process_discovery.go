package api

import (
	"cmp"
	"log"
	"slices"

	"github.com/shirou/gopsutil/v3/process"
)

var processes []ProcInfo

type ProcInfo struct {
	Name string
	CPU  float64
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
		cpu, err := p.CPUPercent()
		name, _ := p.Name()
		if err == nil {
			processes = append(processes, ProcInfo{name, cpu})
		}
	}
}

func sortProcesses(procs []ProcInfo) []ProcInfo {
	slices.SortFunc(procs, func(a, b ProcInfo) int {
		return cmp.Compare(b.CPU, a.CPU)
	})
	return procs
}
