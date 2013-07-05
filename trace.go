package trace

import (
	"strings"
	"sync"
	"time"
)

type Priority uint8

const (
	Trace Priority = iota
	Debug
	Info
	Warn
	Error
)

var lock = new(sync.RWMutex)
var listeners = make([]*Listener, 0)

// Register installs a new Listener
func Register(l *Listener) *ListenerHandle {
	lock.Lock()
	defer lock.Unlock()
	listeners = append(listeners, l)
	return &ListenerHandle{i: len(listeners) - 1, active: true}
}

// M searches for any Listener matching the specified path and
// priority level.  When ok is true the returned match should be
// returned to the library via T or D.
func M(path string, priority Priority) (match []*Listener, ok bool) {
	lock.RLock()
	defer lock.RUnlock()

	if len(listeners) == 0 {
		return
	}

	match = make([]*Listener, 0, len(listeners))
	npath := len(path)

	for _, l := range listeners {
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
func T(match []*Listener, path string, priority Priority, format string, args ...interface{}) {
	if match != nil {
		now := time.Now()
		for _, m := range match {
			m.Fn(now, path, priority, format, args...)
		}
	}
}

// ListenerHandle provides a method to remove a Listener from the registry
type ListenerHandle struct {
	i      int
	active bool
}

// Remove uninstalls a Listener
func (h *ListenerHandle) Remove() {
	lock.Lock()
	defer lock.Unlock()

	if !h.active {
		return
	}
	h.active = false

	n := len(listeners)
	if h.i == 0 {
		if n > 1 {
			listeners = listeners[1:]
		} else {
			listeners = []*Listener{}
		}
	} else if h.i == n {
		listeners = listeners[0:n]
	} else {
		listeners = append(listeners[0:h.i], listeners[h.i+1:]...)
	}
}
