package trace

type Priority uint8

const (
	Trace Priority = iota
	Debug
	Info
	Warn
	Error
)

var priorities = []Priority{Trace, Debug, Info, Warn, Error}

// registry of all added Listener
var registry = NewRegistry()

// Register adds a new Listener
func Register(l *Listener) { registry.Register(l) }

// Remove discards a previously registered Listener
func Remove(l *Listener) { registry.Remove(l) }

// M searches for matching Listener for the given prefix and priority
// level.
func M(p string, n Priority) (match []*Listener, ok bool) {
	return registry.M(p, n)
}

// T calls each match Listener.Fn.
func T(match []*Listener, path string, priority Priority, format string, a ...interface{}) {
	registry.T(match, path, priority, format, a...)
}
