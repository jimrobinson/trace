package trace

import (
	"time"
)

type ListenerFn func(t time.Time, path string, priority Priority, format string, args ...interface{})

type listener struct {
	prefix string   // Call Fn for paths that start with this Prefix
	min    Priority // Call Fn at this Priority or above
	fn     ListenerFn
}

func newListener(prefix string, min Priority, fn ListenerFn) *listener {
	return &listener{
		prefix: prefix,
		min:    min,
		fn:     fn,
	}
}

type listenerState struct {
	path     string
	priority Priority
	fn       ListenerFn
}

func newListenerState(path string, priority Priority, listener *listener) listenerState {
	return listenerState{
		path:     path,
		priority: priority,
		fn:       listener.fn,
	}
}
