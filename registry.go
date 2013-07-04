package trace

import (
	"strings"
	"sync"
	"time"
)

type Registry struct {
	*sync.RWMutex
	set       map[Priority]*Listeners // registered listeners by Priority
	matchBufs chan []*Listener        // cache of []Listener match buffers
}

func NewRegistry() *Registry {
	return &Registry{
		RWMutex:   new(sync.RWMutex),
		set:       make(map[Priority]*Listeners),
		matchBufs: make(chan []*Listener, 100),
	}
}

// Add registers a new Listener
func (r *Registry) Add(l *Listener) {
	for _, n := range l.N {
		r.RLock()
		priority, ok := r.set[n]
		r.RUnlock()

		if ok {
			priority.Lock()
			priority.set = append(priority.set, l)
			priority.Unlock()
		} else {
			r.Lock()
			r.set[n] = NewListeners(l)
			r.Unlock()
		}
	}
}

// Remove discards a previously registered Listener
func (r *Registry) Remove(l *Listener) {
	for _, n := range l.N {
		r.RLock()
		priority, ok := r.set[n]
		r.RUnlock()

		if ok {
			priority.Lock()
			n := len(priority.set)
			for i, t := range priority.set {
				if t == l {
					if i == 0 {
						if n == 1 {
							priority.set = []*Listener{}
						} else {
							priority.set = priority.set[1:]
						}
					} else if i == n {
						priority.set = priority.set[0:n]
					} else {
						priority.set = append(priority.set[0:i], priority.set[i+1:]...)
					}
				}
			}
			priority.Unlock()
		}
	}
}

// M searches for matching Listener for the given prefix and priority
// level.  The returned match is a should be returned to the library
// via T or D.
func (r *Registry) M(p string, n Priority) (match []*Listener, ok bool) {
	r.RLock()
	priority, ok := r.set[n]
	r.RUnlock()

	if ok {
		match = r.popBuf()

		priority.RLock()
		defer priority.RUnlock()

		for _, t := range priority.set {
			if strings.HasPrefix(p, t.P) {
				match = append(match, t)
			}
		}
	}

	return match, len(match) > 0
}

// T calls each match Listener.Fn. This method
// returns control of match to the library.
func (r *Registry) T(match []*Listener, format string, a ...interface{}) {
	defer r.pushBuf(match)
	for _, m := range match {
		m.Fn(time.Now(), format, a...)
	}
}

// D returns control of match to the library.
func (r *Registry) D(match []*Listener) {
	r.pushBuf(match)
}

// popBuf fetches a match buffer if one is available, otherwise it
// allocates a new one.
func (r *Registry) popBuf() []*Listener {
	select {
	case matches := <-r.matchBufs:
		return matches[0:0]
	default:
	}
	return make([]*Listener, 0)
}

// pushBuf returns match to the cache
func (r *Registry) pushBuf(matches []*Listener) {
	select {
	case r.matchBufs <- matches:
	default:
	}
}
