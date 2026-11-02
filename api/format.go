package api

import "fmt"

const (
	kb = 1 << 10
	mb = 1 << 20
	gb = 1 << 30
)

func FormatRAM(bytes float64) string {
	switch {
	case bytes < 10*mb:
		return fmt.Sprintf("%.1f KB", bytes/kb)
	case bytes < gb:
		return fmt.Sprintf("%.1f MB", bytes/mb)
	default:
		return fmt.Sprintf("%.2f GB", bytes/gb)
	}
}

func FormatCPU(percent float64) string {
	return fmt.Sprintf("%.2f%%", percent)
}

// FormatRate renders a throughput in bytes per second, e.g. "1.2 MB/s".
func FormatRate(bytesPerSec float64) string {
	switch {
	case bytesPerSec < kb:
		return fmt.Sprintf("%.0f B/s", bytesPerSec)
	case bytesPerSec < mb:
		return fmt.Sprintf("%.1f KB/s", bytesPerSec/kb)
	case bytesPerSec < gb:
		return fmt.Sprintf("%.1f MB/s", bytesPerSec/mb)
	default:
		return fmt.Sprintf("%.2f GB/s", bytesPerSec/gb)
	}
}

func (p ProcInfo) FormatConsumption(rt ResourceType) string {
	switch rt {
	case RAM:
		return FormatRAM(p.Consumption[rt])
	case CPU:
		return FormatCPU(p.Consumption[rt])
	case NET:
		if !p.Net.Available {
			return "n/a"
		}
		return "↓ " + FormatRate(p.Net.Down) + "  ↑ " + FormatRate(p.Net.Up)
	}
	return ""
}
