package main

import (
	"fmt"
	"linux-system-monitor/api"
	"strings"
	"time"
)

// EXAMPLE of how to use the API:
//  1. api.Start() begins collecting data in the background.
//  2. api.GetSnapshot() returns the latest data; it is cheap, call it as often as you like.
//  3. snapshot.Top(n, rt) returns the n biggest consumers of a resource (api.CPU, api.RAM, api.NET).
func main() {
	rt := api.NET // change to api.CPU or api.RAM to rank by something else

	stop := api.Start(time.Second)
	defer stop()

	const usageWidth = 28
	for {
		s := api.GetSnapshot()
		ramUsed := s.RAMTotal * s.RAMUsedPercent / 100

		fmt.Print("\033[H\033[2J") // clear screen
		fmt.Printf("RAM: %s / %s (%s) | CPU: %s\n",
			api.FormatRAM(ramUsed), api.FormatRAM(s.RAMTotal), api.FormatCPU(s.RAMUsedPercent), api.FormatCPU(s.CPUPercent))
		fmt.Printf("\nTop processes by %s\n", rt)
		fmt.Printf("%-8s %-30s %*s\n", "PID", "NAME", usageWidth, rt)
		fmt.Println(strings.Repeat("-", 8+1+30+1+usageWidth))
		for _, p := range s.Top(10, rt) {
			fmt.Printf("%-8d %-30.30s %*s\n", p.PID, p.Name, usageWidth, p.FormatConsumption(rt))
		}
		time.Sleep(500 * time.Millisecond)
	}
}
