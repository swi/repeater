package main

import (
	"fmt"
	"os"

	"github.com/swi/repeater/pkg/cli"
)

const version = "0.1.0-dev"

func main() {
	config, err := cli.ParseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Handle special cases first
	if config.Help {
		showHelp()
		return
	}

	if config.Version {
		showVersion()
		return
	}

	// Validate configuration
	if err := cli.ValidateConfig(config); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Execute based on subcommand
	switch config.Subcommand {
	case "interval":
		executeInterval(config)
	case "count":
		executeCount(config)
	case "duration":
		executeDuration(config)
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown subcommand: %s\n", config.Subcommand)
		os.Exit(1)
	}
}

func showHelp() {
	fmt.Println("Repeater (rpr) - Continuous Command Execution Tool")
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("  rpr [GLOBAL OPTIONS] <SUBCOMMAND> [OPTIONS] -- <COMMAND>")
	fmt.Println()
	fmt.Println("GLOBAL OPTIONS:")
	fmt.Println("  --help, -h     Show help")
	fmt.Println("  --version, -v  Show version")
	fmt.Println("  --config FILE  Load configuration from file")
	fmt.Println()
	fmt.Println("SUBCOMMANDS:")
	fmt.Println("  interval, int, i    Execute command at regular intervals")
	fmt.Println("  count, cnt, c       Execute command a specific number of times")
	fmt.Println("  duration, dur, d    Execute command for a specific duration")
	fmt.Println()
	fmt.Println("COMMON OPTIONS:")
	fmt.Println("  --every, -e DURATION    Interval between executions")
	fmt.Println("  --times, -t COUNT       Number of times to execute")
	fmt.Println("  --for, -f DURATION      Duration to keep running")
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Println("  rpr interval --every 30s --times 10 -- curl http://example.com")
	fmt.Println("  rpr i -e 30s -t 10 -- curl http://example.com")
	fmt.Println("  rpr count --times 5 -- echo 'Hello World'")
	fmt.Println("  rpr c -t 5 -- echo 'Hello World'")
	fmt.Println("  rpr duration --for 2m --every 10s -- date")
	fmt.Println("  rpr d -f 2m -e 10s -- date")
	fmt.Println()
	fmt.Println("For more information, see: https://github.com/swi/repeater")
}

func showVersion() {
	fmt.Printf("rpr version %s\n", version)
}

func executeInterval(config *cli.Config) {
	fmt.Printf("🕐 Interval execution: every %v", config.Every)
	if config.Times > 0 {
		fmt.Printf(", %d times", config.Times)
	}
	if config.For > 0 {
		fmt.Printf(", for %v", config.For)
	}
	fmt.Printf("\n📋 Command: %v\n", config.Command)
	fmt.Println("⚠️  Scheduler implementation coming next in TDD cycle...")
}

func executeCount(config *cli.Config) {
	fmt.Printf("🔢 Count execution: %d times", config.Times)
	if config.Every > 0 {
		fmt.Printf(", every %v", config.Every)
	}
	fmt.Printf("\n📋 Command: %v\n", config.Command)
	fmt.Println("⚠️  Scheduler implementation coming next in TDD cycle...")
}

func executeDuration(config *cli.Config) {
	fmt.Printf("⏱️  Duration execution: for %v", config.For)
	if config.Every > 0 {
		fmt.Printf(", every %v", config.Every)
	}
	fmt.Printf("\n📋 Command: %v\n", config.Command)
	fmt.Println("⚠️  Scheduler implementation coming next in TDD cycle...")
}
