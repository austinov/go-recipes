package common

import (
	"fmt"
	"os"
	"sync"
	"sync/atomic"
)

type FuseTrigger struct {
	ErrorKind  string
	ErrorLimit int32
	Callback   func(kind string, err error)
}

func NewFuseTrigger(kind string, limit int32, cb func(kind string, err error)) FuseTrigger {
	return FuseTrigger{
		ErrorKind:  kind,
		ErrorLimit: limit,
		Callback:   cb,
	}
}

type state struct {
	FuseTrigger
	errors int32
}

type Fuse struct {
	mu     sync.Mutex
	states map[string]state // key is kind of error
}

func NewFuse(triggers []FuseTrigger) *Fuse {
	states := make(map[string]state)
	for _, t := range triggers {
		states[t.ErrorKind] = state{
			FuseTrigger: t,
		}
	}
	return &Fuse{
		states: states,
	}
}

func (f *Fuse) getState(kind string) (state, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	state, ok := f.states[kind]
	return state, ok
}

func (f *Fuse) setState(kind string, st state) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.states[kind] = st
}

func (f *Fuse) Process(kind string, err error) {
	state, ok := f.getState(kind)
	if !ok {
		fmt.Fprintf(os.Stderr, "Fuse warning: unknown kind of error (%s)\n", kind)
		return
	}
	if err == nil {
		atomic.StoreInt32(&state.errors, 0)
		f.setState(kind, state)
	} else {
		errors := atomic.AddInt32(&state.errors, 1)
		f.setState(kind, state)
		if errors >= state.ErrorLimit {
			state.Callback(kind, err)
		}
	}
}
