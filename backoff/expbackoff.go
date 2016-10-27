package backoff

import (
	"math/rand"
	"time"
)

type ExpBackoff struct {
	initialDelay time.Duration
	maxDelay     time.Duration
	expo         float64
	delay        time.Duration
}

func NewExpBackoff() *ExpBackoff {
	return NewExpBackoffWithConfig(
		1*time.Second,
		10*time.Minute,
		2.0)
}

func NewExpBackoffWithConfig(initDelay, maxDelay time.Duration, expo float64) *ExpBackoff {
	return &ExpBackoff{
		initialDelay: initDelay,
		maxDelay:     maxDelay,
		expo:         expo,
		delay:        initDelay,
	}
}

var rnd = rand.New(rand.NewSource(time.Now().UnixNano()))

func (e *ExpBackoff) Reset() {
	e.delay = e.initialDelay
}

func (e *ExpBackoff) Delay() time.Duration {
	e.delay = e.delay * time.Duration(e.expo)
	if e.delay > e.maxDelay {
		e.delay = e.maxDelay
	}
	normal := time.Duration(rnd.Float64() * 0.1 * float64(time.Second))
	e.delay = e.delay + normal
	return e.delay
}
