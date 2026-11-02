package api

import (
	"errors"
	"time"
)

type NetUsage struct {
	Down      float64
	Up        float64
	Available bool
}

var errNetUnsupported = errors.New("network usage is not supported on this platform")

type netSocket struct {
	PID    int
	Key    string
	Rx, Tx uint64
}

type netTracker struct {
	prev map[string]netSocket
	at   time.Time
}

func (t *netTracker) update(socks []netSocket, now time.Time) map[int]NetUsage {
	next := make(map[string]netSocket, len(socks))
	for _, s := range socks {
		if old, ok := next[s.Key]; ok { // identical keys (e.g. unconnected UDP sockets): merge
			s.Rx += old.Rx
			s.Tx += old.Tx
		}
		next[s.Key] = s
	}

	usage := make(map[int]NetUsage)
	dt := now.Sub(t.at).Seconds()
	if t.prev != nil && dt > 0 {
		for key, s := range next {
			old := t.prev[key] // zero value for a socket opened since the last sample
			u := usage[s.PID]
			u.Down += float64(counterDelta(s.Rx, old.Rx)) / dt
			u.Up += float64(counterDelta(s.Tx, old.Tx)) / dt
			usage[s.PID] = u
		}
	}

	t.prev, t.at = next, now
	return usage
}

func counterDelta(cur, old uint64) uint64 {
	if cur < old {
		return 0
	}
	return cur - old
}
