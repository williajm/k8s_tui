package k8s

import (
	"testing"
	"time"
)

func TestNewExponentialBackoff(t *testing.T) {
	backoff := NewExponentialBackoff()

	if backoff.initialDelay != 1*time.Second {
		t.Errorf("Expected initial delay 1s, got %v", backoff.initialDelay)
	}
	if backoff.maxDelay != 30*time.Second {
		t.Errorf("Expected max delay 30s, got %v", backoff.maxDelay)
	}
	if backoff.multiplier != 2.0 {
		t.Errorf("Expected multiplier 2.0, got %v", backoff.multiplier)
	}
	if backoff.jitter != 0.1 {
		t.Errorf("Expected jitter 0.1, got %v", backoff.jitter)
	}
	if backoff.attempts != 0 {
		t.Errorf("Expected 0 attempts, got %d", backoff.attempts)
	}
}

func TestExponentialBackoffSequence(t *testing.T) {
	backoff := NewExponentialBackoff()

	// Test sequence: 1s, 2s, 4s, 8s, 16s, 30s (max), 30s...
	// Due to jitter (±10%), we check ranges instead of exact values
	expectedSequence := []struct {
		min time.Duration
		max time.Duration
	}{
		{900 * time.Millisecond, 1100 * time.Millisecond},    // ~1s ±10%
		{1800 * time.Millisecond, 2200 * time.Millisecond},   // ~2s ±10%
		{3600 * time.Millisecond, 4400 * time.Millisecond},   // ~4s ±10%
		{7200 * time.Millisecond, 8800 * time.Millisecond},   // ~8s ±10%
		{14400 * time.Millisecond, 17600 * time.Millisecond}, // ~16s ±10%
		{27000 * time.Millisecond, 33000 * time.Millisecond}, // ~30s ±10%
		{27000 * time.Millisecond, 33000 * time.Millisecond}, // ~30s ±10% (capped)
	}

	for i, expected := range expectedSequence {
		delay := backoff.Next()
		if delay < expected.min || delay > expected.max {
			t.Errorf("Attempt %d: expected delay in range [%v, %v], got %v",
				i+1, expected.min, expected.max, delay)
		}
	}

	// Verify attempts counter
	if backoff.Attempts() != len(expectedSequence) {
		t.Errorf("Expected %d attempts, got %d", len(expectedSequence), backoff.Attempts())
	}
}

func TestExponentialBackoffReset(t *testing.T) {
	backoff := NewExponentialBackoff()

	// Generate a few delays
	backoff.Next()
	backoff.Next()
	backoff.Next()

	if backoff.Attempts() != 3 {
		t.Errorf("Expected 3 attempts before reset, got %d", backoff.Attempts())
	}

	// Reset
	backoff.Reset()

	if backoff.Attempts() != 0 {
		t.Errorf("Expected 0 attempts after reset, got %d", backoff.Attempts())
	}

	// First delay after reset should be back to ~1s
	delay := backoff.Next()
	if delay < 900*time.Millisecond || delay > 1100*time.Millisecond {
		t.Errorf("Expected first delay after reset in range [900ms, 1100ms], got %v", delay)
	}
}

func TestExponentialBackoffMaxDelay(t *testing.T) {
	backoff := NewExponentialBackoff()

	// Keep calling Next() until we hit max delay
	var delay time.Duration
	for i := 0; i < 10; i++ {
		delay = backoff.Next()
	}

	// After many attempts, should be capped at ~30s
	if delay < 27*time.Second || delay > 33*time.Second {
		t.Errorf("Expected delay capped at ~30s, got %v", delay)
	}
}

func TestExponentialBackoffJitter(t *testing.T) {
	// Create multiple backoff instances and verify they produce different values
	// due to jitter (random variation)
	delays := make([]time.Duration, 10)
	for i := 0; i < 10; i++ {
		backoff := NewExponentialBackoff()
		delays[i] = backoff.Next()
	}

	// Check that not all delays are identical (jitter is working)
	allSame := true
	first := delays[0]
	for _, d := range delays[1:] {
		if d != first {
			allSame = false
			break
		}
	}

	if allSame {
		t.Errorf("All delays are identical, jitter not working: %v", delays)
	}

	// Verify all delays are in reasonable range for first attempt (~1s ±10%)
	for i, d := range delays {
		if d < 900*time.Millisecond || d > 1100*time.Millisecond {
			t.Errorf("Delay %d out of expected range: %v", i, d)
		}
	}
}

func TestExponentialBackoffCustomConfig(t *testing.T) {
	// Test custom configuration
	backoff := NewExponentialBackoffWithConfig(
		500*time.Millisecond, // initial
		5*time.Second,        // max
		3.0,                  // multiplier
		0.2,                  // jitter (±20%)
	)

	// First delay should be ~500ms ±20%
	delay := backoff.Next()
	if delay < 400*time.Millisecond || delay > 600*time.Millisecond {
		t.Errorf("Expected first delay ~500ms ±20%%, got %v", delay)
	}

	// Second delay should be ~1500ms (500 * 3) ±20%
	delay = backoff.Next()
	if delay < 1200*time.Millisecond || delay > 1800*time.Millisecond {
		t.Errorf("Expected second delay ~1500ms ±20%%, got %v", delay)
	}

	// Third delay should be ~4500ms (1500 * 3) ±20%
	delay = backoff.Next()
	if delay < 3600*time.Millisecond || delay > 5400*time.Millisecond {
		t.Errorf("Expected third delay ~4500ms ±20%%, got %v", delay)
	}

	// Fourth delay should be capped at ~5000ms ±20%
	delay = backoff.Next()
	if delay < 4000*time.Millisecond || delay > 6000*time.Millisecond {
		t.Errorf("Expected fourth delay capped at ~5000ms ±20%%, got %v", delay)
	}
}

func TestExponentialBackoffConcurrency(t *testing.T) {
	backoff := NewExponentialBackoff()
	done := make(chan bool)

	// Simulate concurrent access from multiple goroutines
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 5; j++ {
				_ = backoff.Next()
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should have 50 total attempts (10 goroutines × 5 calls each)
	if backoff.Attempts() != 50 {
		t.Errorf("Expected 50 attempts from concurrent access, got %d", backoff.Attempts())
	}
}

func TestExponentialBackoffZeroJitter(t *testing.T) {
	// Test with zero jitter - delays should be exact
	backoff := NewExponentialBackoffWithConfig(
		1*time.Second,
		30*time.Second,
		2.0,
		0.0, // No jitter
	)

	// First delay should be exactly 1s
	delay := backoff.Next()
	if delay != 1*time.Second {
		t.Errorf("Expected delay 1s (no jitter), got %v", delay)
	}

	// Second delay should be exactly 2s
	delay = backoff.Next()
	if delay != 2*time.Second {
		t.Errorf("Expected delay 2s (no jitter), got %v", delay)
	}

	// Third delay should be exactly 4s
	delay = backoff.Next()
	if delay != 4*time.Second {
		t.Errorf("Expected delay 4s (no jitter), got %v", delay)
	}
}
