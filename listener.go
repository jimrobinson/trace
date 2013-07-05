package trace

import (
	"time"
)

type ListenerFn func(t time.Time, path string, priority Priority, format string, args ...interface{})

type Listener struct {
	Id     string   // Unique identifier for listener
	Prefix string   // Call Fn for paths that start with this Prefix
	Min    Priority // Call Fn at this Priority or above
	Fn     ListenerFn
}

func NewListener(id, prefix string, min Priority, fn ListenerFn) *Listener {
	return &Listener{
		Id:     id,
		Prefix: prefix,
		Min:    min,
		Fn:     fn,
	}
}

type ListenerState struct {
	Path     string
	Priority Priority
	Fn       ListenerFn
}

func NewListenerState(path string, priority Priority, listener *Listener) ListenerState {
	return ListenerState{Path: path, Priority: priority, Fn: listener.Fn}
}
