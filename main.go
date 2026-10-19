package main

import (
	"fmt"
	"linux-system-monitor/api"
	"time"
)

func main() {
	for {
		api.UpdateMonitor()
		fmt.Print("\033[H\033[2J")
		fmt.Print("Top CPU eating processes\n")
		for _, p := range api.GetHeadProcesses(20) {
			fmt.Printf("%-30s %.4f%%\n", p.Name, p.CPU)
		}
		time.Sleep(time.Second)
	}
}
