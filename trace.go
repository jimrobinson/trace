package trace

import (
	"strings"
	"sync"
	"time"
)

var lock = new(sync.RWMutex)
var registry = make([]*listener, 0)

// Register installs a new listener
func Register(prefix string, min Priority, fn ListenerFn) *listenerHandle {
	lock.Lock()
	defer lock.Unlock()
	registry = append(registry, newListener(prefix, min, fn))
	return &listenerHandle{i: len(registry) - 1, active: true}
}

// M searches for any listener matching the specified path and
// priority level.  When ok is true the returned match should be
// returned to the library via T or D.
func M(path string, priority Priority) (match []listenerMatch, ok bool) {
	lock.RLock()
	defer lock.RUnlock()

	if len(registry) == 0 {
		return
	}

	match = make([]listenerMatch, 0, len(registry))
	npath := len(path)

	for _, l := range registry {
		if priority < l.min {
			continue
		}
		if n := len(l.prefix); n > 0 {
			if !strings.HasPrefix(path, l.prefix) {
				continue
			}
			if npath > n && path[n] != '/' {
				continue
			}
		}
		match = append(match, newListenerMatch(path, priority, l))
	}

	return match, len(match) > 0
}

// T logs the format and args to each listener function in match
func T(match []listenerMatch, format string, args ...interface{}) {
	if match != nil {
		now := time.Now()
		for _, m := range match {
			m.fn(now, m.path, m.priority, format, args...)
		}
	}
}

// listenerHandle provides a method to remove a Listener from the registry
type listenerHandle struct {
	i      int
	active bool
}

// Remove uninstalls a listener
func (h *listenerHandle) Remove() {
	lock.Lock()
	defer lock.Unlock()

	if !h.active {
		return
	}
	h.active = false

	n := len(registry)
	if h.i == 0 {
		if n > 1 {
			registry = registry[1:]
		} else {
			registry = []*listener{}
		}
	} else if h.i == n {
		registry = registry[0:n]
	} else {
		registry = append(registry[0:h.i], registry[h.i+1:]...)
	}
}
