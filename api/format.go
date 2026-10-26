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

func (p ProcInfo) FormatConsumption(rt ResourceType) string {
	switch rt {
	case RAM:
		return FormatRAM(p.Consumption[rt])
	case CPU:
		return FormatCPU(p.Consumption[rt])
	}
	return ""
}
