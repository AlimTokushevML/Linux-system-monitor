package main

import (
	"fmt"
	"linux-system-monitor/api"
	"time"
)

func main() {
	for {
		api.UpdateMonitor()
		rt := api.RAM
		procList := api.GetHeadProcesses(10, rt)
		for _, p := range procList {
			fmt.Printf("%-6d %-30s %10s\n", p.PID, p.Name, p.FormatConsumption(rt))

		}
		time.Sleep(time.Second)
	}
}
