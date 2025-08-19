package cli

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// parseSubcommandFlags parses flags specific to subcommands
func (p *argParser) parseSubcommandFlags() error {
	for p.pos < len(p.args) {
		arg := p.args[p.pos]

		if arg == "--" {
			p.pos++ // Skip the separator
			return nil
		}

		switch arg {
		case "--help", "-h":
			p.config.SubcommandHelp = true
			return nil
		case "--every", "-e":
			if err := p.parseDurationFlag(&p.config.Every); err != nil {
				return err
			}
		case "--times", "-t":
			if err := p.parseTimesFlag(); err != nil {
				return err
			}
		case "--for", "-f":
			if err := p.parseDurationFlag(&p.config.For); err != nil {
				return err
			}
		case "--rate", "-r":
			if err := p.parseStringFlag(&p.config.RateSpec); err != nil {
				return err
			}
		case "--retry-pattern", "-p":
			if err := p.parseStringFlag(&p.config.RetryPattern); err != nil {
				return err
			}
		case "--show-next", "-n":
			p.config.ShowNext = true
			p.pos++
		case "--base-interval", "-b":
			if err := p.parseDurationFlag(&p.config.BaseInterval); err != nil {
				return err
			}
		case "--min-interval":
			if err := p.parseDurationFlag(&p.config.MinInterval); err != nil {
				return err
			}
		case "--max-interval":
			if err := p.parseDurationFlag(&p.config.MaxInterval); err != nil {
				return err
			}
		case "--slow-threshold":
			if err := p.parseFloatFlag(&p.config.SlowThreshold); err != nil {
				return err
			}
		case "--fast-threshold":
			if err := p.parseFloatFlag(&p.config.FastThreshold); err != nil {
				return err
			}
		case "--failure-threshold":
			if err := p.parseFloatFlag(&p.config.FailureThreshold); err != nil {
				return err
			}
		case "--show-metrics", "-m":
			p.config.ShowMetrics = true
			p.pos++
		case "--multiplier":
			if err := p.parseFloatFlag(&p.config.Multiplier); err != nil {
				return err
			}
		case "--target-cpu":
			if err := p.parseFloatFlag(&p.config.TargetCPU); err != nil {
				return err
			}
		case "--target-memory":
			if err := p.parseFloatFlag(&p.config.TargetMemory); err != nil {
				return err
			}
		case "--target-load":
			if err := p.parseFloatFlag(&p.config.TargetLoad); err != nil {
				return err
			}
		case "--stream", "-s":
			p.config.Stream = true
			p.pos++
		case "--quiet", "-q":
			p.config.Quiet = true
			p.pos++
		case "--verbose", "-v":
			p.config.Verbose = true
			p.pos++
		case "--stats-only":
			p.config.StatsOnly = true
			p.pos++
		case "--output-prefix", "-o":
			if err := p.parseStringFlag(&p.config.OutputPrefix); err != nil {
				return err
			}
		case "--cron":
			if err := p.parseStringFlag(&p.config.CronExpression); err != nil {
				return err
			}
		case "--timezone", "--tz":
			if err := p.parseStringFlag(&p.config.Timezone); err != nil {
				return err
			}
		case "--success-pattern":
			if err := p.parseStringFlag(&p.config.SuccessPattern); err != nil {
				return err
			}
		case "--failure-pattern":
			if err := p.parseStringFlag(&p.config.FailurePattern); err != nil {
				return err
			}
		case "--case-insensitive":
			p.config.CaseInsensitive = true
			p.pos++
		case "--http-aware":
			p.config.HTTPAware = true
			p.pos++
		case "--http-max-delay":
			if err := p.parseDurationFlag(&p.config.HTTPMaxDelay); err != nil {
				return err
			}
		case "--http-min-delay":
			if err := p.parseDurationFlag(&p.config.HTTPMinDelay); err != nil {
				return err
			}
		case "--http-parse-json":
			p.config.HTTPParseJSON = true
			p.pos++
		case "--http-no-parse-json":
			p.config.HTTPParseJSON = false
			p.pos++
		case "--http-parse-headers":
			p.config.HTTPParseHeaders = true
			p.pos++
		case "--http-no-parse-headers":
			p.config.HTTPParseHeaders = false
			p.pos++
		case "--http-trust-client":
			p.config.HTTPTrustClient = true
			p.pos++
		case "--http-custom-fields":
			if err := p.parseStringSliceFlag(&p.config.HTTPCustomFields); err != nil {
				return err
			}
		case "--attempts", "-a":
			if err := p.parseIntFlag(&p.config.MaxRetries); err != nil {
				return err
			}

		// NEW STRATEGY PARAMETERS
		case "--base-delay", "-bd":
			if err := p.parseDurationFlag(&p.config.BaseDelay); err != nil {
				return err
			}
		case "--increment", "-inc":
			if err := p.parseDurationFlag(&p.config.Increment); err != nil {
				return err
			}
		case "--exponent", "-exp":
			if err := p.parseFloatFlag(&p.config.Exponent); err != nil {
				return err
			}
		case "--max-delay", "-md":
			if err := p.parseDurationFlag(&p.config.MaxDelay); err != nil {
				return err
			}
		// Note: --multiplier already exists at line 356, update it to use new field
		default:
			return fmt.Errorf("unknown flag: %s", arg)
		}
	}
	return nil
}

// parseDurationFlag parses a duration flag value
func (p *argParser) parseDurationFlag(target *time.Duration) error {
	if p.pos+1 >= len(p.args) {
		return fmt.Errorf("%s requires a value", p.args[p.pos])
	}

	duration, err := time.ParseDuration(p.args[p.pos+1])
	if err != nil {
		return fmt.Errorf("invalid duration: %s", p.args[p.pos+1])
	}

	*target = duration
	p.pos += 2
	return nil
}

// parseTimesFlag parses the --times flag value
func (p *argParser) parseTimesFlag() error {
	if p.pos+1 >= len(p.args) {
		return fmt.Errorf("--times requires a value")
	}

	times, err := strconv.ParseInt(p.args[p.pos+1], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid times value: %s", p.args[p.pos+1])
	}

	p.config.Times = times
	p.pos += 2
	return nil
}

// parseStringFlag parses a string flag value
func (p *argParser) parseStringFlag(target *string) error {
	if p.pos+1 >= len(p.args) {
		return fmt.Errorf("%s requires a value", p.args[p.pos])
	}

	*target = p.args[p.pos+1]
	p.pos += 2
	return nil
}

// parseFloatFlag parses a float flag value
func (p *argParser) parseFloatFlag(target *float64) error {
	if p.pos+1 >= len(p.args) {
		return fmt.Errorf("%s requires a value", p.args[p.pos])
	}

	value, err := strconv.ParseFloat(p.args[p.pos+1], 64)
	if err != nil {
		return fmt.Errorf("invalid float value: %s", p.args[p.pos+1])
	}

	*target = value
	p.pos += 2
	return nil
}

// parseIntFlag parses an integer flag value
func (p *argParser) parseIntFlag(target *int) error {
	if p.pos+1 >= len(p.args) {
		return fmt.Errorf("%s requires a value", p.args[p.pos])
	}

	value, err := strconv.Atoi(p.args[p.pos+1])
	if err != nil {
		return fmt.Errorf("invalid integer value: %s", p.args[p.pos+1])
	}

	*target = value
	p.pos += 2
	return nil
}

// parseStringSliceFlag parses a comma-separated string slice flag value
func (p *argParser) parseStringSliceFlag(target *[]string) error {
	if p.pos+1 >= len(p.args) {
		return fmt.Errorf("%s requires a value", p.args[p.pos])
	}

	value := p.args[p.pos+1]
	if value == "" {
		*target = []string{}
	} else {
		*target = strings.Split(value, ",")
		// Trim whitespace from each field
		for i, field := range *target {
			(*target)[i] = strings.TrimSpace(field)
		}
	}

	p.pos += 2
	return nil
}
