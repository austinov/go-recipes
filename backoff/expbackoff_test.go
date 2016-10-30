package backoff

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDelay(t *testing.T) {
	assert := assert.New(t)
	eb := NewExpBackoffWithConfig(Config{
		MinDelay: 100 * time.Millisecond,
		MaxDelay: 5 * time.Second,
		Expo:     2.0,
		Jitter:   0.1,
	})
	assert.NotNil(eb)

	cases := []struct {
		name     string
		expected time.Duration
		reset    bool
	}{
		{
			name:     "1st attempt",
			expected: 100 * time.Millisecond,
		},
		{
			name:     "2nd attempt",
			expected: 200 * time.Millisecond,
		},
		{
			name:     "3rd attempt",
			expected: 400 * time.Millisecond,
		},
		{
			name:     "4th attempt",
			expected: 800 * time.Millisecond,
		},
		{
			name:     "5th attempt",
			expected: 1600 * time.Millisecond,
		},
		{
			name:     "6th attempt",
			expected: 3200 * time.Millisecond,
		},
		{
			name:     "7th attempt",
			expected: 5000 * time.Millisecond,
		},
		{
			name:     "8th attempt",
			expected: 5000 * time.Millisecond,
		},
		{
			name:     "9th attempt",
			expected: 100 * time.Millisecond,
			reset:    true,
		},
		{
			name:     "10th attempt",
			expected: 200 * time.Millisecond,
		},
	}
	for i, c := range cases {
		if c.reset {
			eb.Reset()
		}
		t1 := time.Now()
		t2 := <-eb.Delay()
		delta := math.Ceil(float64(c.expected) * 0.01)
		t.Logf("T%d %v, %v, %v", i+1, c.expected, t2.Sub(t1), c.expected-t2.Sub(t1))
		assert.InDelta(
			float64(c.expected),
			float64(t2.Sub(t1)),
			delta,
			fmt.Sprintf("delay %d attempt ~%v", i+1, c.expected))
	}
}
