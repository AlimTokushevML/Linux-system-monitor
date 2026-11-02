//go:build linux

package api

import (
	"errors"
	"fmt"

	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
)

// readNetSockets reads per-socket TCP byte counters from the kernel through
// netlink sock_diag (the same source as `ss -ti`) and attributes each socket
// to a process via /proc/<pid>/fd.
//
// UDP is not covered: the kernel keeps no byte counters for UDP sockets, so
// UDP-heavy traffic (DNS, QUIC) is invisible here. Sockets of processes we
// can't inspect (other users', without root) are skipped.
func readNetSockets() ([]netSocket, error) {
	owners := scanSocketOwners("/proc")

	var (
		socks    []netSocket
		firstErr error
		anyOK    bool
	)
	for _, family := range []uint8{unix.AF_INET, unix.AF_INET6} {
		resps, err := netlink.SocketDiagTCPInfo(family)
		// an interrupted dump still returns usable (if incomplete) results
		if err != nil && !errors.Is(err, netlink.ErrDumpInterrupted) {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		anyOK = true

		for _, r := range resps {
			if r.InetDiagMsg == nil || r.TCPInfo == nil {
				continue
			}
			pid, ok := owners[r.InetDiagMsg.INode]
			if !ok {
				continue
			}
			socks = append(socks, netSocket{
				PID: pid,
				Key: fmt.Sprintf("tcp|%d", r.InetDiagMsg.INode),
				Rx:  r.TCPInfo.Bytes_received,
				Tx:  r.TCPInfo.Bytes_acked,
			})
		}
	}
	if !anyOK {
		return nil, fmt.Errorf("sock_diag: %w", firstErr)
	}
	return socks, nil
}
