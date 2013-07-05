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

Initially the experiment was not been what I would call a resounding
success, as the overhead involved in splitting T into M and T is a
little under 2x slower than the original:

	github.com/jimrobinson/trace:
	BenchmarkFunctionCall	2000000000	         0.63 ns/op
	BenchmarkNoListeners	20000000	        83.8  ns/op
	BenchmarkOtherListeners	10000000	       217    ns/op
	BenchmarkFirstListener	 5000000	       605    ns/op
	BenchmarkSecondListener	 5000000	       606    ns/op
	BenchmarkBothListeners	 5000000	       615    ns/op

compared to the original library:

	github.com/seehuhn/trace:
	BenchmarkFunctionCall	2000000000	         0.96 ns/op
	BenchmarkNoListeners	20000000	        84.3  ns/op
	BenchmarkOtherListeners	10000000	       188    ns/op
	BenchmarkFirstListener	 5000000	       374    ns/op
	BenchmarkSecondListener	 5000000	       373    ns/op
	BenchmarkBothListeners	 5000000	       381    ns/op

Interestingly enough, changing my implementation to pass the
formatting string and args directly to the listener brings the numbers
back into line with the original implementation:

	github.com/jimrobinson/trace deferred formatting:
	BenchmarkFunctionCall	500000000	         6.71 ns/op
	BenchmarkNoListeners	20000000	        81.5  ns/op
	BenchmarkOtherListeners	10000000	       218    ns/op
	BenchmarkFirstListener	 5000000	       321    ns/op
	BenchmarkSecondListener	 5000000	       323    ns/op
	BenchmarkBothListeners	 5000000	       329    ns/op

But of course that change drops the actual evaluation of the format
and args from the benchmark.  Making the same changes to the original
library brings similar improvements:

	github.com/seehuhn/trace deferred formatting:
	BenchmarkFunctionCall	200000000	         9.45 ns/op
	BenchmarkNoListeners	20000000	        84.3  ns/op
	BenchmarkOtherListeners	10000000	       213    ns/op
	BenchmarkFirstListener	10000000	       206    ns/op
	BenchmarkSecondListener	10000000	       205    ns/op
	BenchmarkBothListeners	10000000	       225    ns/op

I am still hopeful that time and memory savings can be realized in the
case where one is logging large amounts of data, e.g., Trace level
logging of an a large data structure.

Usage
-----

Register a trace.Listener and then call trace.M and trace.T to test
for listener and to call the listeners:

	import "github.com/jimrobinson/trace"

	...
		if m, ok := trace.M(path, priority); ok {
			trace.T(m, path, priority, "%s %d", arg1, arg2)
		}
	...

To install a listener, define a trace.ListenerFn and register it:
        import "log"

        ...
	listenerFn := func(t time.Time, path string, priority trace.Priority, msg string) {
		log.Println(msg)
	}

	listener := NewListener("myListenerId", "a/b/c", trace.Error, listenerFn)

	handle := trace.Register(listener)
        ...

Listeners should be removed when they are no longer needed:

	handle.Remove()

Note that multiple goroutines may call a trace.ListenerFn at a time.
