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

	trace.T(<path>, <priority>, <fmt>, <args>)

into a two-step process:

	if m, ok := trace.M(<path>, <priority>); ok {
		trace.T(m, <path>, <priority>, <fmt>, <args>)
	}

the idea being to not evaluate dynamic args when no listeners were
active for the given path and priority.

So far this experiment has not been what I would call a resounding
success, as the overhead involved in splitting T into M and T is a
little under 2x slower than the original:

       BenchmarkFunctionCall	2000000000	         0.63 ns/op
       BenchmarkNoListeners	20000000	        83.8  ns/op
       BenchmarkOtherListeners	10000000	       217    ns/op
       BenchmarkFirstListener	 5000000	       605    ns/op
       BenchmarkSecondListener	 5000000	       606    ns/op
       BenchmarkBothListeners	 5000000	       615    ns/op

compared to the original library:

	BenchmarkFunctionCall	2000000000	         0.96 ns/op
	BenchmarkNoListeners	20000000	        84.3  ns/op
	BenchmarkOtherListeners	10000000	       188    ns/op
	BenchmarkFirstListener	 5000000	       374    ns/op
	BenchmarkSecondListener	 5000000	       373    ns/op
	BenchmarkBothListeners	 5000000	       381    ns/op

However, I'm not exactly following the structure of the original
library and it may be I'm losing cycles in other parts of the
implementation.

That said, I am still hopeful that time and memory savings can be
realized in the case where one is logging large amounts of data, e.g.,
Trace level logging of an a large data structure.

Usage
-----

Register a trace.Listener and then call trace.M and trace.T to test
for listener and to call the listeners:

	import "github.com/jimrobinson/trace"
	import "log"

	...
		if m, ok := trace.M(path, priority); ok {
			trace.T(m, path, priority, "%s %d", arg1, arg2)
		}
	...

To install a listener, define a trace.ListenerFn and register it:

	listenerFn := func(t time.Time, path string, priority trace.Priority, msg string) {
		log.Println(msg)
	}

	listener := NewListener("myListenerId", "a/b/c", trace.Error, listenerFn)

	handle := trace.Register(listener)

Listeners should be removed when they are no longer needed:

	handle.Remove()

Note that multiple goroutines may call a trace.ListenerFn at a time.
