package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/swi/repeater/pkg/cli"
	"github.com/swi/repeater/pkg/runner"
)

const version = "0.2.0"

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

	// Execute using the integrated runner system
	if err := executeCommand(config); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
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
	fmt.Println("  interval, int, i       Execute command at regular intervals")
	fmt.Println("  count, cnt, c          Execute command a specific number of times")
	fmt.Println("  duration, dur, d       Execute command for a specific duration")
	fmt.Println("  rate-limit, rate, rl   Execute command with server-friendly rate limiting")
	fmt.Println()
	fmt.Println("COMMON OPTIONS:")
	fmt.Println("  --every, -e DURATION       Interval between executions")
	fmt.Println("  --times, -t COUNT          Number of times to execute")
	fmt.Println("  --for, -f DURATION         Duration to keep running")
	fmt.Println("  --rate, -r SPEC            Rate specification (e.g., 10/1h, 100/1m)")
	fmt.Println("  --retry-pattern, -p SPEC   Retry pattern (e.g., 0,10m,30m)")
	fmt.Println("  --show-next, -n            Show next allowed execution time")
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Println("  rpr interval --every 30s --times 10 -- curl http://example.com")
	fmt.Println("  rpr i -e 30s -t 10 -- curl http://example.com")
	fmt.Println("  rpr count --times 5 -- echo 'Hello World'")
	fmt.Println("  rpr c -t 5 -- echo 'Hello World'")
	fmt.Println("  rpr duration --for 2m --every 10s -- date")
	fmt.Println("  rpr d -f 2m -e 10s -- date")
	fmt.Println("  rpr rate-limit --rate 100/1h -- curl https://api.github.com/user")
	fmt.Println("  rpr rl -r 10/1m --retry-pattern 0,5m,15m -- curl api.com")
	fmt.Println("  rpr rate-limit -r 60/1h --show-next -- curl api.example.com")
	fmt.Println()
	fmt.Println("For more information, see: https://github.com/swi/repeater")
}

func showVersion() {
	fmt.Printf("rpr version %s\n", version)
}

func executeCommand(config *cli.Config) error {
	// Create runner
	r, err := runner.NewRunner(config)
	if err != nil {
		return fmt.Errorf("failed to create runner: %w", err)
	}

	// Setup signal handling for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Fprintf(os.Stderr, "\n🛑 Received interrupt signal, shutting down gracefully...\n")
		cancel()
	}()

	// Show execution info
	showExecutionInfo(config)

	// Run the command
	stats, err := r.Run(ctx)
	if err != nil && ctx.Err() == nil {
		// Error that's not from cancellation
		return fmt.Errorf("execution failed: %w", err)
	}

	// Show results
	showExecutionResults(stats)
	return nil
}

func showExecutionInfo(config *cli.Config) {
	switch config.Subcommand {
	case "interval":
		fmt.Printf("🕐 Interval execution: every %v", config.Every)
		if config.Times > 0 {
			fmt.Printf(", %d times", config.Times)
		}
		if config.For > 0 {
			fmt.Printf(", for %v", config.For)
		}
	case "count":
		fmt.Printf("🔢 Count execution: %d times", config.Times)
		if config.Every > 0 {
			fmt.Printf(", every %v", config.Every)
		}
	case "duration":
		fmt.Printf("⏱️  Duration execution: for %v", config.For)
		if config.Every > 0 {
			fmt.Printf(", every %v", config.Every)
		}
	}
	fmt.Printf("\n📋 Command: %v\n", config.Command)
	fmt.Println("🚀 Starting execution...")
}

func showExecutionResults(stats *runner.ExecutionStats) {
	if stats == nil {
		return
	}

	fmt.Printf("\n✅ Execution completed!\n")
	fmt.Printf("📊 Statistics:\n")
	fmt.Printf("   Total executions: %d\n", stats.TotalExecutions)
	fmt.Printf("   Successful: %d\n", stats.SuccessfulExecutions)
	fmt.Printf("   Failed: %d\n", stats.FailedExecutions)
	fmt.Printf("   Duration: %v\n", stats.Duration.Round(time.Millisecond))

	if stats.FailedExecutions > 0 {
		fmt.Printf("\n⚠️  Some executions failed. Check command output above.\n")
	}
}
