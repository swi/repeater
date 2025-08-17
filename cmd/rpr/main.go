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

// ExitError represents an error with a specific exit code
type ExitError struct {
	Code    int
	Message string
}

func (e *ExitError) Error() string {
	return e.Message
}

const version = "0.4.0"

func main() {
	config, err := cli.ParseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(2) // Usage error
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

	// Apply configuration file settings if specified
	if config.ConfigFile != "" {
		if err := applyConfigFile(config); err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config file: %v\n", err)
			os.Exit(2) // Usage error
		}
	}

	// Validate configuration
	if err := cli.ValidateConfig(config); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(2) // Usage error
	}

	// Execute using the integrated runner system
	if err := executeCommand(config); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)

		// Handle different exit codes
		if exitErr, ok := err.(*ExitError); ok {
			os.Exit(exitErr.Code)
		} else {
			os.Exit(1) // General error
		}
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
	fmt.Println("EXECUTION MODES:")
	fmt.Println("  interval, int, i       Execute command at regular intervals")
	fmt.Println("  count, cnt, c          Execute command a specific number of times")
	fmt.Println("  duration, dur, d       Execute command for a specific duration")
	fmt.Println("  cron, cr               Execute command based on cron expressions")
	fmt.Println()
	fmt.Println("MATHEMATICAL RETRY STRATEGIES:")
	fmt.Println("  exponential, exp       Exponential backoff (1s, 2s, 4s, 8s, 16s...)")
	fmt.Println("  fibonacci, fib         Fibonacci backoff (1s, 1s, 2s, 3s, 5s, 8s...)")
	fmt.Println("  linear, lin            Linear backoff (1s, 2s, 3s, 4s, 5s...)")
	fmt.Println("  polynomial, poly       Polynomial backoff with custom exponent")
	fmt.Println("  decorrelated-jitter, dj AWS-recommended distributed retry")
	fmt.Println()
	fmt.Println("ADAPTIVE SCHEDULING:")
	fmt.Println("  adaptive, adapt, a     Execute command with adaptive scheduling")
	fmt.Println("  load-adaptive, load, la Execute command with load-aware adaptive scheduling")
	fmt.Println()
	fmt.Println("RATE CONTROL:")
	fmt.Println("  rate-limit, rate, rl   Execute command with server-friendly rate limiting")
	fmt.Println()
	fmt.Println("LEGACY (DEPRECATED):")
	fmt.Println("  backoff, back, b       Execute command with exponential backoff (use 'exponential')")
	fmt.Println()
	fmt.Println("EXECUTION MODE OPTIONS:")
	fmt.Println("  --every, -e DURATION       Interval between executions")
	fmt.Println("  --times, -t COUNT          Number of times to execute")
	fmt.Println("  --for, -f DURATION         Duration to keep running")
	fmt.Println("  --cron EXPRESSION          Cron expression for scheduling (e.g., '0 9 * * *', '@daily')")
	fmt.Println("  --timezone TZ              Timezone for cron scheduling (default: UTC)")
	fmt.Println()
	fmt.Println("RETRY STRATEGY OPTIONS:")
	fmt.Println("  --base-delay DURATION      Base delay for mathematical strategies (default: 1s)")
	fmt.Println("  --increment DURATION       Linear increment for linear strategy (default: 1s)")
	fmt.Println("  --exponent FLOAT           Polynomial exponent (default: 2.0)")
	fmt.Println("  --multiplier FLOAT         Growth multiplier for exponential/jitter (default: 2.0)")
	fmt.Println("  --max-delay DURATION       Maximum delay cap for all strategies (default: 60s)")
	fmt.Println("  --attempts, -a COUNT       Maximum retry attempts (default: 3)")
	fmt.Println()
	fmt.Println("ADAPTIVE SCHEDULING OPTIONS:")
	fmt.Println("  --base-interval, -b DUR    Base interval for adaptive scheduling")
	fmt.Println("  --show-metrics, -m         Show adaptive scheduling metrics")
	fmt.Println("  --target-cpu FLOAT         Target CPU usage % for load-adaptive (default: 70)")
	fmt.Println("  --target-memory FLOAT      Target memory usage % for load-adaptive (default: 80)")
	fmt.Println("  --target-load FLOAT        Target load average for load-adaptive (default: 1.0)")
	fmt.Println()
	fmt.Println("RATE CONTROL OPTIONS:")
	fmt.Println("  --rate, -r SPEC            Rate specification (e.g., 10/1h, 100/1m)")
	fmt.Println("  --retry-pattern, -p SPEC   Retry pattern (e.g., 0,10m,30m)")
	fmt.Println("  --show-next, -n            Show next allowed execution time")
	fmt.Println()
	fmt.Println("LEGACY OPTIONS (DEPRECATED):")
	fmt.Println("  --initial-delay, -i DUR    Initial interval for backoff (use --base-delay)")
	fmt.Println("  --max, -x DUR              Maximum backoff interval (use --max-delay)")
	fmt.Println("  --jitter FLOAT             Jitter factor 0.0-1.0 (default: 0.0)")
	fmt.Println()
	fmt.Println("OUTPUT CONTROL:")
	fmt.Println("  --quiet, -q                Suppress command output, show only tool errors")
	fmt.Println("  --verbose, -v              Show detailed execution info + command output")
	fmt.Println("  --stats-only               Show only execution statistics")
	fmt.Println("  --stream, -s               Force streaming output (default for pipeline mode)")
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Println("  # Basic usage")
	fmt.Println("  rpr interval --every 30s --times 10 -- curl http://example.com")
	fmt.Println("  rpr i -e 30s -t 10 -- curl http://example.com")
	fmt.Println("  rpr count --times 5 -- echo 'Hello World'")
	fmt.Println("  rpr c -t 5 -- echo 'Hello World'")
	fmt.Println("  rpr duration --for 2m --every 10s -- date")
	fmt.Println("  rpr d -f 2m -e 10s -- date")
	fmt.Println()
	fmt.Println("  # Unix pipeline integration")
	fmt.Println("  rpr i -e 5s -t 10 -- curl -s https://api.com | jq .status")
	fmt.Println("  rpr c -t 20 -- curl -w \"%{time_total}\\n\" -s -o /dev/null api.com | sort -n")
	fmt.Println("  rpr d -f 1h -e 5m -- df -h / | awk 'NR==2{print $5}' | tee disk.log")
	fmt.Println()
	fmt.Println("  # Mathematical retry strategies")
	fmt.Println("  rpr exponential --base-delay 1s --attempts 5 -- curl flaky-api.com")
	fmt.Println("  rpr exp --base-delay 500ms --max-delay 30s --attempts 3 -- ping google.com")
	fmt.Println("  rpr fibonacci --base-delay 1s --attempts 8 -- curl api.com")
	fmt.Println("  rpr linear --increment 2s --attempts 5 -- ./retry-script.sh")
	fmt.Println("  rpr polynomial --base-delay 1s --exponent 1.5 --attempts 4 -- command")
	fmt.Println()
	fmt.Println("  # Advanced scheduling")
	fmt.Println("  rpr rate-limit --rate 100/1h -- curl https://api.github.com/user")
	fmt.Println("  rpr adaptive --base-interval 1s --show-metrics -- curl api.com")
	fmt.Println("  rpr load-adaptive --base-interval 1s --target-cpu 70 -- ./task.sh")
	fmt.Println("  rpr cron --cron '0 9 * * *' -- ./daily-backup.sh  # Every day at 9 AM")
	fmt.Println("  rpr cron --cron '@hourly' --timezone America/New_York -- curl api.com")
	fmt.Println()
	fmt.Println("  # Output modes")
	fmt.Println("  rpr i -e 5s -t 3 --quiet -- curl https://api.com  # Silent")
	fmt.Println("  rpr i -e 5s -t 3 --verbose -- curl https://api.com  # Detailed")
	fmt.Println("  rpr i -e 5s -t 3 --stats-only -- curl https://api.com  # Stats only")
	fmt.Println()
	fmt.Println("EXIT CODES:")
	fmt.Println("  0   All commands succeeded")
	fmt.Println("  1   Some commands failed")
	fmt.Println("  2   Usage error")
	fmt.Println("  130 Interrupted (Ctrl+C)")
	fmt.Println()
	fmt.Println("For more information, see: https://github.com/swi/repeater")
}

func showVersion() {
	fmt.Printf("rpr version %s\n", version)
}

func executeCommand(config *cli.Config) error {
	// Apply Unix pipeline behavior: make streaming default unless quiet or stats-only mode
	if !config.Quiet && !config.StatsOnly && !config.Stream {
		config.Stream = true
	}

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
		if config.Verbose {
			fmt.Fprintf(os.Stderr, "\nReceived interrupt signal, shutting down gracefully...\n")
		}
		cancel()
	}()

	// Unix pipeline behavior: only show execution info in verbose mode
	if config.Verbose {
		showExecutionInfo(config)
	}

	// Run the command
	stats, err := r.Run(ctx)
	if err != nil {
		if ctx.Err() == context.Canceled {
			// Interrupted by signal (Ctrl+C)
			return &ExitError{Code: 130, Message: "interrupted"}
		}
		// Other execution errors
		return &ExitError{Code: 1, Message: fmt.Sprintf("execution failed: %v", err)}
	}

	// Unix pipeline behavior: show results in verbose or stats-only mode
	if config.Verbose || config.StatsOnly {
		showExecutionResults(stats)
	}

	// Check if any commands failed
	if stats != nil && stats.FailedExecutions > 0 {
		return &ExitError{Code: 1, Message: "some commands failed"}
	}

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
	case "adaptive":
		fmt.Printf("🧠 Adaptive execution: base interval %v", config.BaseInterval)
		if config.MinInterval > 0 {
			fmt.Printf(", range %v-%v", config.MinInterval, config.MaxInterval)
		}
		if config.ShowMetrics {
			fmt.Printf(", with metrics")
		}
	case "backoff":
		fmt.Fprintf(os.Stderr, "⚠️  Warning: 'backoff' is deprecated, use 'exponential' instead\n")
		fmt.Printf("📈 Exponential backoff: initial %v", config.InitialInterval)
		if config.BackoffMax > 0 {
			fmt.Printf(", max %v", config.BackoffMax)
		}
		if config.BackoffMultiplier > 0 {
			fmt.Printf(", multiplier %.1fx", config.BackoffMultiplier)
		}
		if config.BackoffJitter > 0 {
			fmt.Printf(", jitter %.1f%%", config.BackoffJitter*100)
		}
	case "exponential":
		fmt.Printf("📈 Exponential strategy: base delay %v", config.BaseDelay)
		if config.MaxDelay > 0 {
			fmt.Printf(", max %v", config.MaxDelay)
		}
		if config.Multiplier > 0 {
			fmt.Printf(", multiplier %.1fx", config.Multiplier)
		}
	case "fibonacci":
		fmt.Printf("🌀 Fibonacci strategy: base delay %v", config.BaseDelay)
		if config.MaxDelay > 0 {
			fmt.Printf(", max %v", config.MaxDelay)
		}
	case "linear":
		fmt.Printf("📏 Linear strategy: increment %v", config.Increment)
		if config.MaxDelay > 0 {
			fmt.Printf(", max %v", config.MaxDelay)
		}
	case "polynomial":
		fmt.Printf("🔢 Polynomial strategy: base delay %v", config.BaseDelay)
		if config.Exponent > 0 {
			fmt.Printf(", exponent %.1f", config.Exponent)
		}
		if config.MaxDelay > 0 {
			fmt.Printf(", max %v", config.MaxDelay)
		}
	case "decorrelated-jitter":
		fmt.Printf("🎲 Decorrelated jitter strategy: base delay %v", config.BaseDelay)
		if config.Multiplier > 0 {
			fmt.Printf(", multiplier %.1fx", config.Multiplier)
		}
		if config.MaxDelay > 0 {
			fmt.Printf(", max %v", config.MaxDelay)
		}
	case "load-adaptive":
		fmt.Printf("⚖️  Load-adaptive execution: base interval %v", config.BaseInterval)
		if config.TargetCPU > 0 {
			fmt.Printf(", target CPU %.0f%%", config.TargetCPU)
		}
		if config.TargetMemory > 0 {
			fmt.Printf(", memory %.0f%%", config.TargetMemory)
		}
		if config.TargetLoad > 0 {
			fmt.Printf(", load %.1f", config.TargetLoad)
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
