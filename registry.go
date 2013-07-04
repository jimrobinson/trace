package trace

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

type Registry struct {
	*sync.RWMutex
	listeners []*Listener
	matchBufs chan []*Listener
}

func NewRegistry() *Registry {
	return &Registry{
		RWMutex:   new(sync.RWMutex),
		listeners: make([]*Listener, 0),
		matchBufs: make(chan []*Listener, 100),
	}
}

// Register installs a new Listener
func (r *Registry) Register(l *Listener) {
	r.Lock()
	r.listeners = append(r.listeners, l)
	r.Unlock()
}

// Remove uninstalls the specified Listener
func (r *Registry) Remove(l *Listener) {
	r.Lock()
	n := len(r.listeners)
	for i, v := range r.listeners {
		if v.Id != l.Id {
			continue
		}
		if i == 0 {
			if n > 1 {
				r.listeners = r.listeners[1:]
			} else {
				r.listeners = []*Listener{}
			}
		} else if i == n {
			r.listeners = r.listeners[0:n]
		} else {
			r.listeners = append(r.listeners[0:i], r.listeners[i+1:]...)
		}
	}
	r.Unlock()
}

// M searches for any Listener matching the specified path and
// priority level.  When ok is true the returned match should be
// returned to the library via T or D.
func (r *Registry) M(path string, priority Priority) (match []*Listener, ok bool) {
	r.RLock()
	defer r.RUnlock()

	if len(r.listeners) == 0 {
		return
	}

	match = make([]*Listener, 0, len(r.listeners))
	npath := len(path)

	for _, l := range r.listeners {
		if priority < l.Min {
			continue
		}
		if n := len(l.Prefix); n > 0 {
			if !strings.HasPrefix(path, l.Prefix) {
				continue
			}
			if npath > n && path[n] != '/' {
				continue
			}
		}
		match = append(match, l)
	}

	return match, len(match) > 0
}

// T calls each match Listener.Fn.
func (r *Registry) T(match []*Listener, path string, priority Priority, format string, args ...interface{}) {
	if match != nil {
		now := time.Now()
		msg := fmt.Sprintf(format, args...)
		for _, m := range match {
			m.Fn(now, path, priority, msg)
		}
	}
}
