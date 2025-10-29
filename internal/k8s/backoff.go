package k8s

import (
	"math"
	"math/rand"
	"sync"
	"time"
)

// ExponentialBackoff implements exponential backoff with jitter for reconnection delays.
// It provides progressive delay between reconnection attempts to avoid overwhelming
// the API server while allowing quick recovery from transient failures.
type ExponentialBackoff struct {
	initialDelay time.Duration
	maxDelay     time.Duration
	multiplier   float64
	jitter       float64
	attempts     int
	mu           sync.Mutex
}

// NewExponentialBackoff creates a new exponential backoff with sensible defaults:
// - Initial delay: 1 second
// - Max delay: 30 seconds
// - Multiplier: 2.0 (doubles each attempt)
// - Jitter: 0.1 (±10% randomization)
func NewExponentialBackoff() *ExponentialBackoff {
	return &ExponentialBackoff{
		initialDelay: 1 * time.Second,
		maxDelay:     30 * time.Second,
		multiplier:   2.0,
		jitter:       0.1,
		attempts:     0,
	}
}

// NewExponentialBackoffWithConfig creates a backoff with custom parameters.
func NewExponentialBackoffWithConfig(initial, max time.Duration, multiplier, jitter float64) *ExponentialBackoff {
	return &ExponentialBackoff{
		initialDelay: initial,
		maxDelay:     max,
		multiplier:   multiplier,
		jitter:       jitter,
		attempts:     0,
	}
}

// Next calculates and returns the next backoff duration.
// The delay follows the sequence: 1s, 2s, 4s, 8s, 16s, 30s (max), 30s...
// Each delay includes random jitter to avoid thundering herd.
func (eb *ExponentialBackoff) Next() time.Duration {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	eb.attempts++

	// Calculate exponential delay
	delay := float64(eb.initialDelay) * math.Pow(eb.multiplier, float64(eb.attempts-1))

	// Cap at maximum delay
	if delay > float64(eb.maxDelay) {
		delay = float64(eb.maxDelay)
	}

	// Add jitter: ±10% randomization
	jitterRange := delay * eb.jitter
	jitterValue := (rand.Float64() * 2 * jitterRange) - jitterRange
	delay += jitterValue

	// Ensure minimum of 0
	if delay < 0 {
		delay = 0
	}

	return time.Duration(delay)
}

// Reset resets the backoff to initial state.
// Should be called after a successful connection.
func (eb *ExponentialBackoff) Reset() {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	eb.attempts = 0
}

// Attempts returns the current number of attempts.
func (eb *ExponentialBackoff) Attempts() int {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	return eb.attempts
}
