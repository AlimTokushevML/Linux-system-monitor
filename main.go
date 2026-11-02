package main

import (
	"fmt"
	"linux-system-monitor/api"
	"strings"
	"time"
)

func main() {

	// EXAMPLE of how to use the API
	//
	rt := api.CPU
	title := "RAM"
	if rt == api.CPU {
		title = "CPU"
	}

	stop := api.Start(time.Second)
	defer stop()

	for {
		s := api.GetSnapshot()
		ramUsed := s.RAMTotal * s.RAMUsedPercent / 100

		fmt.Print("\033[H\033[2J")
		fmt.Printf("RAM: %s / %s (%s) | CPU: %s\n",
			api.FormatRAM(ramUsed), api.FormatRAM(s.RAMTotal), api.FormatCPU(s.RAMUsedPercent), api.FormatCPU(s.CPUPercent))
		fmt.Printf("\nTop processes by %s\n", title)
		fmt.Printf("%-8s %-30s %12s\n", "PID", "NAME", title)
		fmt.Println(strings.Repeat("-", 52))
		for _, p := range s.Top(10, rt) {
			fmt.Printf("%-8d %-30.30s %12s\n", p.PID, p.Name, p.FormatConsumption(rt))
		}
		time.Sleep(500 * time.Millisecond)
	}
}
