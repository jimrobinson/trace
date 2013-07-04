package trace

import (
	"sync"
	"time"
)

type Listeners struct {
	*sync.RWMutex
	set []*Listener // registered listeners
}

func NewListeners(set ...*Listener) *Listeners {
	return &Listeners{
		RWMutex: new(sync.RWMutex),
		set:     set,
	}
}

type Listener struct {
	Fn func(time.Time, string, ...interface{})
	P  string
	N  []Priority
}

func NewListener(fn func(time.Time, string, ...interface{}), p string, n ...Priority) *Listener {
	return &Listener{Fn: fn, P: p, N: n}
}
