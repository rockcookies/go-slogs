package slogs

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSystemClock(t *testing.T) {
	t.Run("Now returns current time", func(t *testing.T) {
		clock := systemClock{}

		before := time.Now()
		now := clock.Now()
		after := time.Now()

		assert.True(t, now.After(before) || now.Equal(before), "Now() should return time after or equal to before")
		assert.True(t, now.Before(after) || now.Equal(after), "Now() should return time before or equal to after")
	})

	t.Run("Now returns monotonic increasing values", func(t *testing.T) {
		clock := systemClock{}

		t1 := clock.Now()
		t2 := clock.Now()

		assert.True(t, t2.After(t1) || t2.Equal(t1), "Subsequent calls to Now() should be monotonic")
	})

	t.Run("NewTicker creates functional ticker", func(t *testing.T) {
		clock := systemClock{}
		duration := 10 * time.Millisecond

		ticker := clock.NewTicker(duration)
		require.NotNil(t, ticker, "NewTicker should return a non-nil ticker")
		defer ticker.Stop()

		select {
		case <-ticker.C:
			// First tick received, ticker is working
		case <-time.After(duration * 2):
			t.Fatal("Expected to receive a tick within 2x duration")
		}
	})

	t.Run("NewTicker respects duration", func(t *testing.T) {
		clock := systemClock{}
		duration := 5 * time.Millisecond

		ticker := clock.NewTicker(duration)
		defer ticker.Stop()

		start := time.Now()
		select {
		case <-ticker.C:
			elapsed := time.Since(start)
			assert.True(t, elapsed >= duration, "Tick should fire after at least the specified duration")
			assert.True(t, elapsed < duration*3, "Tick should fire within reasonable time (3x duration)")
		case <-time.After(duration * 5):
			t.Fatal("Expected to receive a tick within 5x duration")
		}
	})
}

func TestDefaultClock(t *testing.T) {
	t.Run("DefaultClock is systemClock", func(t *testing.T) {
		assert.IsType(t, systemClock{}, DefaultClock, "DefaultClock should be of type systemClock")
	})

	t.Run("DefaultClock Now works", func(t *testing.T) {
		now := DefaultClock.Now()
		assert.False(t, now.IsZero(), "DefaultClock.Now() should return non-zero time")
	})

	t.Run("DefaultClock NewTicker works", func(t *testing.T) {
		ticker := DefaultClock.NewTicker(1 * time.Millisecond)
		assert.NotNil(t, ticker, "DefaultClock.NewTicker should return a non-nil ticker")
		ticker.Stop()
	})
}

func TestClockInterface(t *testing.T) {
	t.Run("systemClock implements Clock interface", func(t *testing.T) {
		var _ Clock = systemClock{}
		var _ Clock = &systemClock{}
	})
}

// TestClockUsage demonstrates how to use Clock in practice
func TestClockUsage(t *testing.T) {
	t.Run("Clock can be used for time-dependent operations", func(t *testing.T) {
		clock := systemClock{}

		// Simulate a time-dependent operation
		start := clock.Now()

		// Simulate some work
		time.Sleep(1 * time.Millisecond)

		duration := clock.Now().Sub(start)
		assert.GreaterOrEqual(t, duration, time.Millisecond, "Operation should take at least 1ms")
	})
}

// BenchmarkClockNow benchmarks the Now() method
func BenchmarkClockNow(b *testing.B) {
	clock := systemClock{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = clock.Now()
	}
}

// BenchmarkNewTicker benchmarks the NewTicker() method
func BenchmarkNewTicker(b *testing.B) {
	clock := systemClock{}
	duration := time.Millisecond

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ticker := clock.NewTicker(duration)
		ticker.Stop()
	}
}