package main

import (
	"fmt"
	"linux-system-monitor/api"
	"time"
)

func main() {
	for {
		var rt api.ResourceType
		rt = api.CPU
		api.UpdateMonitor()
		fmt.Print("\033[H\033[2J")
		fmt.Print("Top Resource Eating processes\n")
		fmt.Print("PID    |      NAME       | USAGE\n")
		for _, p := range api.GetHeadProcesses(20, rt) {
			if rt == api.CPU {
				fmt.Printf("%-6d %-30s %8.4f%%\n", p.PID, p.Name, p.Consumption[rt])
			} else if rt == api.RAM {
				fmt.Printf("%-6d %-30s %8.3f GiB\n", p.PID, p.Name, p.Consumption[rt]/(1<<30))
			}

		}
		time.Sleep(time.Second)
	}
}
