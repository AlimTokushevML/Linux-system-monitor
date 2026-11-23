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
	NET     // download + upload, bytes per second
	NETDown // download, bytes per second
	NETUp   // upload, bytes per second
)

func (rt ResourceType) String() string {
	switch rt {
	case CPU:
		return "CPU"
	case RAM:
		return "RAM"
	case NET:
		return "NET"
	case NETDown:
		return "NET DOWN"
	case NETUp:
		return "NET UP"
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
			procs[i].Consumption[NETDown] = u.Down
			procs[i].Consumption[NETUp] = u.Up
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

// Sorted returns all processes ordered by their usage of rt: highest first
// when desc is true, lowest first otherwise. Ties are ordered by PID so the
// list doesn't jitter between refreshes. It returns a new slice and does not
// modify the snapshot.
func (s *Snapshot) Sorted(rt ResourceType, desc bool) []ProcInfo {
	sorted := slices.Clone(s.Processes)
	slices.SortFunc(sorted, func(a, b ProcInfo) int {
		byUsage := cmp.Compare(a.Consumption[rt], b.Consumption[rt])
		if desc {
			byUsage = -byUsage
		}
		return cmp.Or(byUsage, cmp.Compare(a.PID, b.PID))
	})
	return sorted
}

// Top returns up to n processes with the highest usage of rt, highest first.
func (s *Snapshot) Top(n int, rt ResourceType) []ProcInfo {
	sorted := s.Sorted(rt, true)
	return sorted[:max(0, min(n, len(sorted)))]
}
