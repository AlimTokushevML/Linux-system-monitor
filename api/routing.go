package api

import (
	"cmp"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

type ResourceType int

const (
	CPU ResourceType = iota
	RAM
	NET
)

func (rt ResourceType) String() string {
	switch rt {
	case CPU:
		return "CPU"
	case RAM:
		return "RAM"
	case NET:
		return "NET"
	}
	return "unknown"
}

type Snapshot struct {
	CPUPercent     float64
	RAMUsedPercent float64
	RAMTotal       float64
	Processes      []ProcInfo
	Taken          time.Time
}

var current atomic.Pointer[Snapshot]

var (
	updateMu sync.Mutex
	netTrack netTracker
)

func updateMonitor() error {
	updateMu.Lock()
	defer updateMu.Unlock()

	procs, err := findProcesses()
	if err != nil {
		return err
	}
	cpuPercent, err := cpu.Percent(time.Second, false)
	if err != nil {
		return err
	}
	vm, err := mem.VirtualMemory()
	if err != nil {
		return err
	}

	if socks, err := readNetSockets(); err == nil {
		usage := netTrack.update(socks, time.Now())
		for i := range procs {
			u := usage[procs[i].PID]
			u.Available = true
			procs[i].Net = u
			procs[i].Consumption[NET] = u.Down + u.Up
		}
	}

	current.Store(&Snapshot{
		CPUPercent:     cpuPercent[0],
		RAMUsedPercent: vm.UsedPercent,
		RAMTotal:       float64(vm.Total),
		Processes:      procs,
		Taken:          time.Now(),
	})
	return nil
}

func Start(interval time.Duration) (stop func()) {
	_ = updateMonitor()

	done := make(chan struct{})
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				_ = updateMonitor()
			}
		}
	}()
	return func() { close(done) }
}

func GetSnapshot() *Snapshot {
	if s := current.Load(); s != nil {
		return s
	}
	return &Snapshot{}
}

func (s *Snapshot) Top(n int, rt ResourceType) []ProcInfo {
	sorted := slices.Clone(s.Processes)
	slices.SortFunc(sorted, func(a, b ProcInfo) int {
		return cmp.Compare(b.Consumption[rt], a.Consumption[rt])
	})
	return sorted[:max(0, min(n, len(sorted)))]
}
