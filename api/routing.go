package api

import (
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
)

func GetTotalCPU() float64 {
	percentage, err := cpu.Percent(time.Second, false)
	if err != nil {
		panic("Failed to get CPU usage")
	}
	return percentage[0]
}
