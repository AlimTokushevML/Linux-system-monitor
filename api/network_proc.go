package api

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// scanSocketOwners maps socket inode -> PID by reading <procRoot>/<pid>/fd/*.
// A socket shared by several processes (after fork) goes to the first PID
// found. Processes whose fds can't be read (other users', without root) are
// skipped, so their sockets stay unowned.
func scanSocketOwners(procRoot string) map[uint32]int {
	owners := make(map[uint32]int)
	entries, err := os.ReadDir(procRoot)
	if err != nil {
		return owners
	}
	for _, e := range entries {
		pid, err := strconv.Atoi(e.Name())
		if err != nil {
			continue
		}
		fdDir := filepath.Join(procRoot, e.Name(), "fd")
		fds, err := os.ReadDir(fdDir)
		if err != nil {
			continue
		}
		for _, fd := range fds {
			link, err := os.Readlink(filepath.Join(fdDir, fd.Name()))
			if err != nil {
				continue
			}
			if inode, ok := parseSocketLink(link); ok {
				if _, taken := owners[inode]; !taken {
					owners[inode] = pid
				}
			}
		}
	}
	return owners
}

// parseSocketLink extracts the inode from an fd link such as "socket:[12345]".
func parseSocketLink(link string) (uint32, bool) {
	rest, ok := strings.CutPrefix(link, "socket:[")
	if !ok {
		return 0, false
	}
	rest, ok = strings.CutSuffix(rest, "]")
	if !ok {
		return 0, false
	}
	n, err := strconv.ParseUint(rest, 10, 32)
	return uint32(n), err == nil
}
