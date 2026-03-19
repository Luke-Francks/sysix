package main

import (
	"fmt"
	"os"

	"github.com/Luke-Francks/sysix/internal/collector"
)

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		printHelp()
		return
	}

	switch args[0] {
	case "status":
		showProcs := false
		showPorts := false
		for _, arg := range args[1:] {
			if arg == "--procs" {
				showProcs = true
			}
			if arg == "--ports" {
				showPorts = true
			}
		}

		snapshot, err := collector.GetSnapshot()
		if err != nil {
			fmt.Println("error collecting system data:", err)
			return
		}
		fmt.Printf("\n[*] Host:   %s (%s)\n", snapshot.Hostname, snapshot.OS)
		fmt.Printf("[~] Uptime: %d hours\n", snapshot.Uptime/3600)
		fmt.Printf("[>] CPU:    %.1f%%\n", snapshot.CPUPercent)
		fmt.Printf("[>] Memory: %.1f%% (%d MB / %d MB)\n", snapshot.MemPercent, snapshot.MemUsed/1024/1024, snapshot.MemTotal/1024/1024)
		fmt.Printf("[>] Disk:   %.1f%% (%d GB / %d GB)\n", snapshot.DiskPercent, snapshot.DiskUsed/1024/1024/1024, snapshot.DiskTotal/1024/1024/1024)

		if showProcs {
			fmt.Println("\n--- Processes ---")
			fmt.Printf("%-6s %-30s %8s %10s %s\n", "PID", "Name", "CPU%", "Mem(MB)", "Status")
			fmt.Println("----------------------------------------------------------------------")
			for _, p := range snapshot.Processes {
				if p.MemMB > 1 {
					fmt.Printf("%-6d %-30s %8.1f %10.1f %s\n", p.PID, p.Name, p.CPUPercent, p.MemMB, p.Status)
				}
			}
		}

		if showPorts {
			ports, err := collector.GetPorts()
			if err != nil {
				fmt.Println("error collecting ports:", err)
			} else {
				fmt.Println("\n--- Open Ports ---")
				fmt.Printf("%-8s %-6s %-8s %s\n", "Port", "Type", "Status", "PID")
				fmt.Println("--------------------------------")
				for _, p := range ports {
					fmt.Printf("%-8d %-6s %-8s %d\n", p.Port, p.Type, p.Status, p.PID)
				}
			}
		}

	case "watch":
		fmt.Println("sysix watch (TUI) - not yet implemented")
	case "serve":
		fmt.Println("sysix serve (web UI) - not yet implemented")
	case "help":
		printHelp()
	default:
		fmt.Printf("unknown command: %s\n", args[0])
		printHelp()
	}
}

func printHelp() {
	fmt.Println(`
welcome to sysix observer 0.3

Usage:
  sysix status            snapshot of your system right now
  sysix status --procs    include running processes
  sysix status --ports    include open ports
  sysix watch             launch the live TUI
  sysix serve             launch the web UI
  sysix help              show this help message
	`)
}
