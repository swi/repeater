package adaptive

import (
	"errors"
	"sync"
	"time"
)

// CircuitState represents the state of a circuit breaker
type CircuitState int

const (
	CircuitClosed CircuitState = iota
	CircuitOpen
	CircuitHalfOpen
)

// ExecutionResult represents the result of a command execution
type ExecutionResult struct {
	Timestamp    time.Time
	ResponseTime time.Duration
	Success      bool
	StatusCode   int
	Error        error
}

// AdaptiveConfig holds configuration for adaptive scheduling
type AdaptiveConfig struct {
	// AIMD parameters
	BaseInterval           time.Duration `json:"base_interval"`
	MinInterval            time.Duration `json:"min_interval"`
	MaxInterval            time.Duration `json:"max_interval"`
	AdditiveIncrease       time.Duration `json:"additive_increase"`
	MultiplicativeDecrease float64       `json:"multiplicative_decrease"`

	// EWMA parameters
	ResponseTimeAlpha   float64 `json:"response_time_alpha"`
	SlowThresholdFactor float64 `json:"slow_threshold_factor"`
	FastThresholdFactor float64 `json:"fast_threshold_factor"`

	// Bayesian parameters
	PriorAlpha        float64 `json:"prior_alpha"`
	PriorBeta         float64 `json:"prior_beta"`
	DecayRate         float64 `json:"decay_rate"`
	FailureThreshold  float64 `json:"failure_threshold"`
	RecoveryThreshold float64 `json:"recovery_threshold"`

	// Learning parameters
	WindowSize int `json:"window_size"`
	MinSamples int `json:"min_samples"`
}

// AIMDConfig holds configuration for AIMD adapter
type AIMDConfig struct {
	BaseInterval           time.Duration
	MinInterval            time.Duration
	MaxInterval            time.Duration
	AdditiveIncrease       time.Duration
	MultiplicativeDecrease float64
	EWMAAlpha              float64
	SlowThresholdFactor    float64
	FastThresholdFactor    float64
}

// BayesianConfig holds configuration for Bayesian predictor
type BayesianConfig struct {
	PriorAlpha        float64
	PriorBeta         float64
	DecayRate         float64
	FailureThreshold  float64
	RecoveryThreshold float64
}

// AdaptiveMetrics holds metrics for adaptive scheduling
type AdaptiveMetrics struct {
	TotalExecutions      int64
	SuccessfulExecutions int64
	FailedExecutions     int64
	AverageResponseTime  time.Duration
	CurrentInterval      time.Duration
	SuccessRate          float64
	CircuitState         CircuitState
}

// AIMDAdapter implements Additive Increase/Multiplicative Decrease adaptation
type AIMDAdapter struct {
	mu                     sync.RWMutex
	currentInterval        time.Duration
	baseInterval           time.Duration
	maxInterval            time.Duration
	minInterval            time.Duration
	additiveIncrease       time.Duration
	multiplicativeDecrease float64
	avgResponseTime        time.Duration
	ewmaAlpha              float64
	slowThreshold          time.Duration
	fastThreshold          time.Duration
}

// NewAIMDAdapter creates a new AIMD adapter
func NewAIMDAdapter(config *AIMDConfig) *AIMDAdapter {
	return &AIMDAdapter{
		currentInterval:        config.BaseInterval,
		baseInterval:           config.BaseInterval,
		maxInterval:            config.MaxInterval,
		minInterval:            config.MinInterval,
		additiveIncrease:       config.AdditiveIncrease,
		multiplicativeDecrease: config.MultiplicativeDecrease,
		ewmaAlpha:              config.EWMAAlpha,
		slowThreshold:          time.Duration(float64(config.BaseInterval) * config.SlowThresholdFactor),
		fastThreshold:          time.Duration(float64(config.BaseInterval) * config.FastThresholdFactor),
	}
}

// UpdateInterval updates the current interval based on response time and success
func (a *AIMDAdapter) UpdateInterval(responseTime time.Duration, success bool) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Update EWMA of response times
	if a.avgResponseTime == 0 {
		a.avgResponseTime = responseTime
	} else {
		a.avgResponseTime = time.Duration(
			float64(a.avgResponseTime)*(1-a.ewmaAlpha) +
				float64(responseTime)*a.ewmaAlpha,
		)
	}

	if !success {
		// Multiplicative decrease on failure
		a.currentInterval = time.Duration(
			float64(a.currentInterval) * a.multiplicativeDecrease,
		)
	} else if a.avgResponseTime > a.slowThreshold {
		// Additive increase if responses are slow
		a.currentInterval += a.additiveIncrease
	} else if a.avgResponseTime < a.fastThreshold {
		// Additive decrease if responses are fast
		if a.currentInterval > a.additiveIncrease {
			a.currentInterval -= a.additiveIncrease
		}
	}

	// Clamp to bounds
	if a.currentInterval < a.minInterval {
		a.currentInterval = a.minInterval
	}
	if a.currentInterval > a.maxInterval {
		a.currentInterval = a.maxInterval
	}
}

// GetCurrentInterval returns the current interval
func (a *AIMDAdapter) GetCurrentInterval() time.Duration {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.currentInterval
}

// GetAverageResponseTime returns the current average response time
func (a *AIMDAdapter) GetAverageResponseTime() time.Duration {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.avgResponseTime
}

// BayesianPredictor implements Bayesian success prediction with circuit breaker
type BayesianPredictor struct {
	mu                sync.RWMutex
	alpha             float64
	beta              float64
	decayRate         float64
	failureThreshold  float64
	recoveryThreshold float64
	state             CircuitState
}

// NewBayesianPredictor creates a new Bayesian predictor
func NewBayesianPredictor(config *BayesianConfig) *BayesianPredictor {
	return &BayesianPredictor{
		alpha:             config.PriorAlpha,
		beta:              config.PriorBeta,
		decayRate:         config.DecayRate,
		failureThreshold:  config.FailureThreshold,
		recoveryThreshold: config.RecoveryThreshold,
		state:             CircuitClosed,
	}
}

// UpdatePattern updates the pattern with a new success/failure observation
func (b *BayesianPredictor) UpdatePattern(success bool) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Apply exponential decay to existing counts
	b.alpha *= b.decayRate
	b.beta *= b.decayRate

	// Add new observation
	if success {
		b.alpha += 1.0
	} else {
		b.beta += 1.0
	}

	// Update circuit breaker state
	successRate := b.alpha / (b.alpha + b.beta)

	switch b.state {
	case CircuitClosed:
		if successRate < b.failureThreshold {
			b.state = CircuitOpen
		}
	case CircuitOpen:
		if successRate > b.recoveryThreshold {
			b.state = CircuitHalfOpen
		}
	case CircuitHalfOpen:
		if successRate > b.recoveryThreshold {
			b.state = CircuitClosed
		} else if successRate < b.failureThreshold {
			b.state = CircuitOpen
		}
	}
}

// GetSuccessProbability returns the current success probability
func (b *BayesianPredictor) GetSuccessProbability() float64 {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.alpha / (b.alpha + b.beta)
}

// GetCircuitState returns the current circuit breaker state
func (b *BayesianPredictor) GetCircuitState() CircuitState {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.state
}

// AdaptiveScheduler combines AIMD and Bayesian prediction for adaptive scheduling
type AdaptiveScheduler struct {
	mu                sync.RWMutex
	aimdAdapter       *AIMDAdapter
	bayesianPredictor *BayesianPredictor
	config            *AdaptiveConfig
	recentExecutions  []ExecutionResult
	metrics           *AdaptiveMetrics
}

// NewAdaptiveScheduler creates a new adaptive scheduler
func NewAdaptiveScheduler(config *AdaptiveConfig) *AdaptiveScheduler {
	aimdConfig := &AIMDConfig{
		BaseInterval:           config.BaseInterval,
		MinInterval:            config.MinInterval,
		MaxInterval:            config.MaxInterval,
		AdditiveIncrease:       config.AdditiveIncrease,
		MultiplicativeDecrease: config.MultiplicativeDecrease,
		EWMAAlpha:              config.ResponseTimeAlpha,
		SlowThresholdFactor:    config.SlowThresholdFactor,
		FastThresholdFactor:    config.FastThresholdFactor,
	}

	bayesianConfig := &BayesianConfig{
		PriorAlpha:        config.PriorAlpha,
		PriorBeta:         config.PriorBeta,
		DecayRate:         config.DecayRate,
		FailureThreshold:  config.FailureThreshold,
		RecoveryThreshold: config.RecoveryThreshold,
	}

	return &AdaptiveScheduler{
		aimdAdapter:       NewAIMDAdapter(aimdConfig),
		bayesianPredictor: NewBayesianPredictor(bayesianConfig),
		config:            config,
		recentExecutions:  make([]ExecutionResult, 0, config.WindowSize),
		metrics: &AdaptiveMetrics{
			CircuitState: CircuitClosed,
		},
	}
}

// NewAdaptiveSchedulerWithValidation creates a new adaptive scheduler with config validation
func NewAdaptiveSchedulerWithValidation(config *AdaptiveConfig) (*AdaptiveScheduler, error) {
	if err := validateAdaptiveConfig(config); err != nil {
		return nil, err
	}
	return NewAdaptiveScheduler(config), nil
}

// UpdateFromResult updates the scheduler based on an execution result
func (a *AdaptiveScheduler) UpdateFromResult(result ExecutionResult) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Update AIMD adapter
	a.aimdAdapter.UpdateInterval(result.ResponseTime, result.Success)

	// Update Bayesian predictor
	a.bayesianPredictor.UpdatePattern(result.Success)

	// Store for pattern analysis
	a.recentExecutions = append(a.recentExecutions, result)
	if len(a.recentExecutions) > a.config.WindowSize {
		a.recentExecutions = a.recentExecutions[1:]
	}

	// Update metrics
	a.updateMetrics(result)
}

// GetCurrentInterval returns the current adapted interval
func (a *AdaptiveScheduler) GetCurrentInterval() time.Duration {
	a.mu.RLock()
	defer a.mu.RUnlock()

	baseInterval := a.aimdAdapter.GetCurrentInterval()

	// Apply circuit breaker logic
	switch a.bayesianPredictor.GetCircuitState() {
	case CircuitOpen:
		// Exponential backoff when circuit is open
		backoffInterval := baseInterval * 4
		if backoffInterval > a.config.MaxInterval {
			return a.config.MaxInterval
		}
		return backoffInterval
	case CircuitHalfOpen:
		// Conservative interval when testing recovery
		conservativeInterval := baseInterval * 2
		if conservativeInterval > a.config.MaxInterval {
			return a.config.MaxInterval
		}
		return conservativeInterval
	default:
		return baseInterval
	}
}

// GetSuccessProbability returns the current success probability
func (a *AdaptiveScheduler) GetSuccessProbability() float64 {
	return a.bayesianPredictor.GetSuccessProbability()
}

// GetMetrics returns current adaptive scheduling metrics
func (a *AdaptiveScheduler) GetMetrics() *AdaptiveMetrics {
	a.mu.RLock()
	defer a.mu.RUnlock()

	// Create a copy to avoid race conditions
	metrics := *a.metrics
	metrics.CurrentInterval = a.aimdAdapter.GetCurrentInterval()
	metrics.AverageResponseTime = a.aimdAdapter.GetAverageResponseTime()
	metrics.SuccessRate = a.bayesianPredictor.GetSuccessProbability()
	metrics.CircuitState = a.bayesianPredictor.GetCircuitState()

	return &metrics
}

// updateMetrics updates internal metrics (must be called with lock held)
func (a *AdaptiveScheduler) updateMetrics(result ExecutionResult) {
	a.metrics.TotalExecutions++
	if result.Success {
		a.metrics.SuccessfulExecutions++
	} else {
		a.metrics.FailedExecutions++
	}
}

// validateAdaptiveConfig validates the adaptive configuration
func validateAdaptiveConfig(config *AdaptiveConfig) error {
	if config.MinInterval >= config.MaxInterval {
		return errors.New("min_interval must be less than max_interval")
	}

	if config.BaseInterval < config.MinInterval || config.BaseInterval > config.MaxInterval {
		return errors.New("base_interval must be between min_interval and max_interval")
	}

	if config.ResponseTimeAlpha < 0 || config.ResponseTimeAlpha > 1 {
		return errors.New("response_time_alpha must be between 0 and 1")
	}

	if config.FailureThreshold < 0 || config.FailureThreshold > 1 {
		return errors.New("failure_threshold must be between 0 and 1")
	}

	if config.RecoveryThreshold < 0 || config.RecoveryThreshold > 1 {
		return errors.New("recovery_threshold must be between 0 and 1")
	}

	if config.DecayRate < 0 || config.DecayRate > 1 {
		return errors.New("decay_rate must be between 0 and 1")
	}

	if config.WindowSize <= 0 {
		return errors.New("window_size must be positive")
	}

	return nil
}

// DefaultAdaptiveConfig returns a default adaptive configuration
func DefaultAdaptiveConfig() *AdaptiveConfig {
	return &AdaptiveConfig{
		BaseInterval:           time.Second,
		MinInterval:            100 * time.Millisecond,
		MaxInterval:            30 * time.Second,
		AdditiveIncrease:       200 * time.Millisecond,
		MultiplicativeDecrease: 0.6,
		ResponseTimeAlpha:      0.1,
		SlowThresholdFactor:    2.0,
		FastThresholdFactor:    0.5,
		PriorAlpha:             1.0,
		PriorBeta:              1.0,
		DecayRate:              0.95,
		FailureThreshold:       0.3,
		RecoveryThreshold:      0.8,
		WindowSize:             100,
		MinSamples:             10,
	}
}

// DefaultAIMDConfig returns a default AIMD configuration
func DefaultAIMDConfig() *AIMDConfig {
	return &AIMDConfig{
		BaseInterval:           time.Second,
		MinInterval:            100 * time.Millisecond,
		MaxInterval:            30 * time.Second,
		AdditiveIncrease:       200 * time.Millisecond,
		MultiplicativeDecrease: 0.6,
		EWMAAlpha:              0.1,
		SlowThresholdFactor:    2.0,
		FastThresholdFactor:    0.5,
	}
}

// DefaultBayesianConfig returns a default Bayesian configuration
func DefaultBayesianConfig() *BayesianConfig {
	return &BayesianConfig{
		PriorAlpha:        1.0,
		PriorBeta:         1.0,
		DecayRate:         0.95,
		FailureThreshold:  0.3,
		RecoveryThreshold: 0.8,
	}
}
