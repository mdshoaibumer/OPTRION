package circuitbreaker

import (
	"errors"
	"sync"
	"time"
)

// State represents the circuit breaker state.
type State int

const (
	StateClosed   State = iota // Normal operation — requests pass through
	StateOpen                  // Failure threshold exceeded — requests rejected immediately
	StateHalfOpen              // Testing recovery — limited requests allowed
)

func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// ErrCircuitOpen is returned when the circuit breaker is in the open state.
var ErrCircuitOpen = errors.New("circuit breaker is open")

// Config holds circuit breaker configuration.
type Config struct {
	// MaxFailures is the number of consecutive failures before opening the circuit.
	MaxFailures int
	// Timeout is how long the circuit stays open before transitioning to half-open.
	Timeout time.Duration
	// HalfOpenMaxRequests is the number of requests allowed in half-open state.
	HalfOpenMaxRequests int
}

// DefaultConfig returns sensible defaults for AI provider calls.
func DefaultConfig() Config {
	return Config{
		MaxFailures:         3,
		Timeout:             30 * time.Second,
		HalfOpenMaxRequests: 1,
	}
}

// CircuitBreaker implements the circuit breaker pattern.
type CircuitBreaker struct {
	mu              sync.Mutex
	config          Config
	state           State
	failures        int
	successes       int
	lastFailureTime time.Time
	halfOpenCount   int
}

// New creates a new circuit breaker with the given configuration.
func New(cfg Config) *CircuitBreaker {
	if cfg.MaxFailures <= 0 {
		cfg.MaxFailures = 3
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 30 * time.Second
	}
	if cfg.HalfOpenMaxRequests <= 0 {
		cfg.HalfOpenMaxRequests = 1
	}
	return &CircuitBreaker{
		config: cfg,
		state:  StateClosed,
	}
}

// Execute runs the given function through the circuit breaker.
// Returns ErrCircuitOpen if the circuit is open.
func (cb *CircuitBreaker) Execute(fn func() error) error {
	if !cb.allowRequest() {
		return ErrCircuitOpen
	}

	err := fn()

	if err != nil {
		cb.recordFailure()
		return err
	}

	cb.recordSuccess()
	return nil
}

// State returns the current state of the circuit breaker.
func (cb *CircuitBreaker) State() State {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.currentState()
}

// Reset manually resets the circuit breaker to closed state.
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.state = StateClosed
	cb.failures = 0
	cb.successes = 0
	cb.halfOpenCount = 0
}

func (cb *CircuitBreaker) allowRequest() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.currentState() {
	case StateClosed:
		return true
	case StateOpen:
		return false
	case StateHalfOpen:
		if cb.halfOpenCount < cb.config.HalfOpenMaxRequests {
			cb.halfOpenCount++
			return true
		}
		return false
	}
	return false
}

func (cb *CircuitBreaker) currentState() State {
	if cb.state == StateOpen {
		// Check if timeout has elapsed — transition to half-open
		if time.Since(cb.lastFailureTime) >= cb.config.Timeout {
			cb.state = StateHalfOpen
			cb.halfOpenCount = 0
			cb.successes = 0
		}
	}
	return cb.state
}

func (cb *CircuitBreaker) recordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		cb.failures = 0
	case StateHalfOpen:
		cb.successes++
		if cb.successes >= cb.config.HalfOpenMaxRequests {
			// Recovery confirmed — close the circuit
			cb.state = StateClosed
			cb.failures = 0
			cb.successes = 0
			cb.halfOpenCount = 0
		}
	}
}

func (cb *CircuitBreaker) recordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures++
	cb.lastFailureTime = time.Now()

	switch cb.state {
	case StateClosed:
		if cb.failures >= cb.config.MaxFailures {
			cb.state = StateOpen
		}
	case StateHalfOpen:
		// Any failure in half-open re-opens the circuit
		cb.state = StateOpen
		cb.halfOpenCount = 0
		cb.successes = 0
	}
}
