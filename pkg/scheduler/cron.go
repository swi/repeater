package scheduler

import (
	"fmt"
	"time"

	"github.com/swi/repeater/pkg/cron"
)

// Scheduler interface for type safety
type Scheduler interface {
	Next() <-chan time.Time
	Stop()
}

// CronScheduler implements Scheduler using cron expressions
type CronScheduler struct {
	expression *cron.CronExpression
	timezone   *time.Location
	nextChan   chan time.Time
	stopChan   chan struct{}
	stopped    bool
}

// NewCronScheduler creates a new cron scheduler
func NewCronScheduler(expression, timezone string) (*CronScheduler, error) {
	// Parse the cron expression
	cronExpr, err := cron.ParseCron(expression)
	if err != nil {
		return nil, fmt.Errorf("invalid cron expression: %w", err)
	}

	// Parse the timezone
	tz, err := time.LoadLocation(timezone)
	if err != nil {
		return nil, fmt.Errorf("invalid timezone: %w", err)
	}

	return &CronScheduler{
		expression: cronExpr,
		timezone:   tz,
		nextChan:   make(chan time.Time, 1),
		stopChan:   make(chan struct{}),
		stopped:    false,
	}, nil
}

// Next returns a channel that will receive the next execution time
func (c *CronScheduler) Next() <-chan time.Time {
	go c.schedule()
	return c.nextChan
}

// Stop stops the scheduler
func (c *CronScheduler) Stop() {
	if !c.stopped {
		c.stopped = true
		close(c.stopChan)
	}
}

// schedule runs the scheduling logic in a goroutine
func (c *CronScheduler) schedule() {
	for {
		// Calculate next execution time
		now := time.Now().In(c.timezone)
		next := c.expression.NextExecution(now)

		// Wait until the next execution time
		waitDuration := next.Sub(now)
		if waitDuration <= 0 {
			// If the time has already passed, schedule for the next occurrence
			next = c.expression.NextExecution(next)
			waitDuration = next.Sub(now)
		}

		select {
		case <-time.After(waitDuration):
			// Time to execute
			select {
			case c.nextChan <- next:
				// Successfully sent the execution time
			case <-c.stopChan:
				// Scheduler was stopped while trying to send
				return
			}
		case <-c.stopChan:
			// Scheduler was stopped while waiting
			return
		}
	}
}
