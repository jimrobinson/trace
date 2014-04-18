Trace
=====

A simple tracing framework based on github.com/seehuhn/trace.

Overview
--------

Jochen Voss posted a nice library for calling trace functions from
within other programs, allowing for interested parties to register
Listeners that could then act upon those logging calls.

I was curious what the overhead would be to split out what was a one
call method:

	trace.T(<listenerFn>, <fmt>, <args>)

into a two-step process:

	traceFn, traceT := trace.M(<path>, trace.Trace);
	
	...

	if traceT {
		trace.T(traceFn, <fmt>, <args>)
	}

where 'traceFn/traceT' might be 'errorFn/errorT' or
'warnFn/warnT' depending on the trace level.

the idea being to not evaluate dynamic args when no listeners were
active for the given path and priority.

Initially the experiment was not been what I would call a resounding
success, as the overhead involved in splitting T into M and T is
around 2x slower than the original:

	github.com/jimrobinson/trace:
	BenchmarkFunctionCall	500000000	         6.64 ns/op
	BenchmarkNoListeners	20000000	        82.7 ns/op
	BenchmarkOtherListeners	10000000	       247 ns/op
	BenchmarkFirstListener	 5000000	       499 ns/op
	BenchmarkSecondListener	 5000000	       500 ns/op
	BenchmarkBothListeners	 5000000	       569 ns/op

Note that this implementation defers evaluation of the format and args
to the Listener, raising the cost of having multiple Listeners in both
memory and cpu cycles.

I am still hopeful that time and memory savings can be realized in the
case where one is logging large amounts of data, e.g., Trace level
logging of an a large data structure.

Usage
-----

Register a trace.Listener and then call trace.M and trace.T to test
for listeners and to call those listeners:

	import "github.com/jimrobinson/trace"

	...
		traceId := "github.com/jimrobinson/xml/xmlbase"
	...
		traceFn, traceT := trace.M(traceId, trace.Trace)
		infoFn, infoT := trace.M(traceId, trace.Info)
	...
		if traceT {
			trace.T(traceFn, "got %s %d", arg1, arg2)
		}
	...
		if infoT {
			trace.T(infoFn, "got %s %d", arg1, arg2)
		}
	...

To install a listener, define a trace.ListenerFn and register it:
        import "log"

        ...
	listenerFn := func(t time.Time, path string, priority trace.Priority, msg string) {
		log.Println(msg)
	}

	listener := NewListener("myListenerId", traceId, trace.Info, listenerFn)

	handle := trace.Register(listener)
        ...

Listeners should be removed when they are no longer needed:

	handle.Remove()

Note that multiple goroutines may call a trace.ListenerFn at a time.
