package backoff

import (
	"math/rand"
	"time"
)

type Config struct {
	MinDelay time.Duration
	MaxDelay time.Duration
	Expo     float64
	Jitter   float64
}

type ExpBackoff struct {
	config   Config
	delay    time.Duration
	attempts uint64
}

var (
	DefaultConfig = Config{
		100 * time.Millisecond,
		10 * time.Minute,
		2.0,
		0.1,
	}
	rnd = rand.New(rand.NewSource(time.Now().UnixNano()))
)

func NewExpBackoff() *ExpBackoff {
	return NewExpBackoffWithConfig(DefaultConfig)
}

func NewExpBackoffWithConfig(config Config) *ExpBackoff {
	return &ExpBackoff{
		config: config,
		delay:  config.MinDelay,
	}
}

func (e *ExpBackoff) Reset() {
	e.delay = e.config.MinDelay
	e.attempts = 0
}

func (e *ExpBackoff) Delay() <-chan time.Time {
	e.delay = e.delay * time.Duration(e.config.Expo)
	if e.delay > e.config.MaxDelay {
		e.delay = e.config.MaxDelay
	}
	normal := time.Duration(rnd.Float64() * e.config.Jitter * float64(time.Second))
	e.delay = e.delay + normal
	return time.After(e.delay)
}
