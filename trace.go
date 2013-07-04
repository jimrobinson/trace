package trace

type Priority uint8

const (
	Error Priority = iota
	Warn
	Info
	Debug
	Trace
)

var All = []Priority{Error, Warn, Info, Debug, Trace}

// registry of all added Listener
var registry = NewRegistry()

// Add registers a new Listener
func Add(l *Listener) { registry.Add(l) }

// Remove discards a previously registered Listener
func Remove(l *Listener) { registry.Remove(l) }

// M searches for matching Listener for the given prefix and priority
// level.  The returned match is a cache structure that should be
// returned to the library via T or D.
func M(p string, n Priority) (match []*Listener, ok bool) { return registry.M(p, n) }

// T calls each match Listener.Fn. This method
// returns control of match to the library.
func T(match []*Listener, format string, a ...interface{}) { registry.T(match, format, a...) }

// D returns control of match to the library.
func D(match []*Listener) { registry.D(match) }
