package main

import (
	"sync"
	"github.com/jwijenbergh/puregotk/v4/gobject"
)

// SignalPool is a collection of functions that will be called to cleanup all signals
type SignalPool struct {
	mu sync.Mutex
	cleanup []func()
}

// AddSignal adds a signal for a puregotk object
// It does this by adding a function that constructs a gobject from the pointer
// And then disconnects the signal `t`
func (p *SignalPool) AddSignal(ptr gobject.Ptr, t uint32) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.cleanup = append(p.cleanup, func() {
		var obj gobject.Object
		obj.SetGoPointer(ptr.GoPointer())
		obj.DisconnectSignal(t)
	})
}

// DisconnectSignals loops over the whole collection and calls the cleanup handler functions
func (p *SignalPool) DisconnectSignals() {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, v := range p.cleanup {
		v()
	}
}
