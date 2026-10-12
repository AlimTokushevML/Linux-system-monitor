package api

import (
	"log"
	"sort"

	"github.com/shirou/gopsutil/v3/process"
)

var processes []ProcInfo

type ProcInfo struct {
	Name string
	CPU  float64
}

func findProcesses() {
	procs, err := process.Processes()
	if err != nil {
		log.Fatal("Failed to retrieve the processes")
	}

	for _, p := range procs {
		cpu, err := p.CPUPercent()
		name, _ := p.Name()
		if err == nil {
			processes = append(processes, ProcInfo{name, cpu})
		}
	}
}

func sortProcesses() {
	sort.Slice(processes, func(i, j int) bool {
		return processes[i].CPU > processes[j].CPU
	})
}
