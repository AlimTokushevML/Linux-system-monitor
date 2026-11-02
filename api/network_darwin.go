//go:build darwin

package api

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var pidSuffix = regexp.MustCompile(`:(\d+)$`)

// readNetSockets reads per-socket byte counters from `netstat -anv`, which
// reports rxbytes, txbytes and the owning process for TCP and UDP sockets.
func readNetSockets() ([]netSocket, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()
	var socks []netSocket
	for _, proto := range []string{"tcp", "udp"} {
		out, err := exec.CommandContext(ctx, "netstat", "-anv", "-p", proto).Output()
		if err != nil {
			return nil, fmt.Errorf("netstat %s: %w", proto, err)
		}
		socks = append(socks, parseNetstat(string(out))...)
	}
	return socks, nil
}

func parseNetstat(out string) []netSocket {
	var socks []netSocket
	for _, line := range strings.Split(out, "\n") {
		f := strings.Fields(line)
		if len(f) < 10 {
			continue
		}

		var rxCol int
		switch {
		case strings.HasPrefix(f[0], "tcp"):
			rxCol = 6
		case strings.HasPrefix(f[0], "udp"):
			rxCol = 5
		default:
			continue
		}
		rx, err := strconv.ParseUint(f[rxCol], 10, 64)
		if err != nil {
			continue
		}
		tx, err := strconv.ParseUint(f[rxCol+1], 10, 64)
		if err != nil {
			continue
		}

		pid := -1
		for i := len(f) - 1; i > rxCol+1; i-- {
			if m := pidSuffix.FindStringSubmatch(f[i]); m != nil {
				pid, _ = strconv.Atoi(m[1])
				break
			}
		}
		if pid < 0 {
			continue
		}

		socks = append(socks, netSocket{
			PID: pid,
			Key: fmt.Sprintf("%s|%s|%s|%d", f[0], f[3], f[4], pid),
			Rx:  rx,
			Tx:  tx,
		})
	}
	return socks
}
