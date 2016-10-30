package backoff

import (
	"math/rand"
	"sync"
	"time"
)

// Config defines the config for ExpBackoff
type Config struct {
	MinDelay time.Duration
	MaxDelay time.Duration
	Expo     float64
	Jitter   float64
}

// ExpBackoff is a thread-safe implemention of
// the backoff algorithm to exponential increase
// the delay between repeated processes
// in the case of unsuccessful attempts.
type ExpBackoff struct {
	config Config

	// this mutex protects delay and attempts
	mu       sync.Mutex
	delay    time.Duration
	attempts uint64
}

var (
	// DefaultConfig is the default ExpBackoff config.
	DefaultConfig = Config{
		100 * time.Millisecond,
		1 * time.Minute,
		2.0,
		0.1,
	}
	rnd = rand.New(rand.NewSource(time.Now().UnixNano()))
)

// NewExpBackoff returns an ExpBackoff with default configuration.
func NewExpBackoff() *ExpBackoff {
	return NewExpBackoffWithConfig(DefaultConfig)
}

// NewExpBackoffWithConfig returns an ExpBackoff by passed configuration.
func NewExpBackoffWithConfig(config Config) *ExpBackoff {
	if config.MaxDelay == 0 {
		config.MaxDelay = DefaultConfig.MaxDelay
	}
	if config.Expo == 0 {
		config.Expo = DefaultConfig.Expo
	}
	return &ExpBackoff{
		config: config,
		delay:  config.MinDelay,
	}
}

// Delay waits for exponentially increased duration and send the current time
// on the returned channel.
func (e *ExpBackoff) Delay() <-chan time.Time {
	e.mu.Lock()
	currDelay := e.delay
	nextDelay := currDelay
	e.mu.Unlock()

	nextDelay *= time.Duration(e.config.Expo)
	if nextDelay > e.config.MaxDelay {
		nextDelay = e.config.MaxDelay
	}
	normal := time.Duration(rnd.NormFloat64() * e.config.Jitter * float64(time.Millisecond))
	nextDelay += normal

	e.mu.Lock()
	e.delay = nextDelay
	e.attempts++
	e.mu.Unlock()
	return time.After(currDelay)
}

// Reset sets current delay the minimum value
func (e *ExpBackoff) Reset() {
	e.mu.Lock()
	e.attempts = 0
	e.delay = e.config.MinDelay
	e.mu.Unlock()
}

// Attempts returns number of unsuccessful attempts
func (e *ExpBackoff) Attempts() uint64 {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.attempts
}
