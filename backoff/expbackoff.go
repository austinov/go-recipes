package backoff

import (
	"math/rand"
	"time"
)

type Config struct {
	initialDelay time.Duration
	maxDelay     time.Duration
	expo         float64
}

type ExpBackoff struct {
	config Config
	delay  time.Duration
}

var (
	DefaultConfig = Config{
		100 * time.Millisecond,
		600000 * time.Millisecond,
		2.0,
	}
	rnd = rand.New(rand.NewSource(time.Now().UnixNano()))
)

func NewExpBackoff() *ExpBackoff {
	return NewExpBackoffWithConfig(DefaultConfig)
}

func NewExpBackoffWithConfig(config Config) *ExpBackoff {
	return &ExpBackoff{
		config: config,
		delay:  config.initialDelay,
	}
}

func (e *ExpBackoff) Reset() {
	e.delay = e.config.initialDelay
}

func (e *ExpBackoff) Delay() time.Duration {
	e.delay = e.delay * time.Duration(e.config.expo)
	if e.delay > e.config.maxDelay {
		e.delay = e.config.maxDelay
	}
	normal := time.Duration(rnd.Float64() * 0.1 * float64(time.Second))
	e.delay = e.delay + normal
	return e.delay
}
