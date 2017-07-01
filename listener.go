package trace

import (
	"fmt"
	"time"
)

// ListenerFn defines a function used to register a trace listener.
// Variable t specifies the time that the trace event was created,
// path specifies an identifier (e..g github.com/jimrobinson/trace),
// priority specifies the minimum priority level accepted.  The format
// and args values specify an fmt.Sprintf compatible message.
type ListenerFn func(t time.Time, path string, priority Priority, format string, args ...interface{})

// FormatterFn defines a function used to format a trace event into
// a string.
type FormatterFn func(t time.Time, path string, priority Priority, format string, args ...interface{}) string

// DefaultFormaterFn defines a default FormatterFn.  It will produce
// a message format "[<time>][<path>] <message>", where time uses the
// format time.RFC3339.
var DefaultFormatterFn = func(t time.Time, id string, priority Priority, format string, args ...interface{}) string {
	return fmt.Sprintf("[%s][%s] %s", t.Format(time.RFC3339), id, fmt.Sprintf(format, args...))
}

// listener defines a ListenerFn that should be called when a trace
// path starts with prefix and when it has a Priority level >= min.
type listener struct {
	prefix string
	min    Priority
	fn     ListenerFn
}

func newListener(prefix string, min Priority, fn ListenerFn) *listener {
	return &listener{
		prefix: prefix,
		min:    min,
		fn:     fn,
	}
}

// listenerMatch is produced by function M and is used to
// identify ListenerFn from the registry that are interested
// in a message.
type listenerMatch struct {
	path     string
	priority Priority
	fn       ListenerFn
}

func newListenerMatch(path string, priority Priority, listener *listener) listenerMatch {
	return listenerMatch{
		path:     path,
		priority: priority,
		fn:       listener.fn,
	}
}
