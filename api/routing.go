package api

import (
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
)

func getRAM() {
	cpu.Percent(time.Second, false)
}
