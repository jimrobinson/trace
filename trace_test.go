package trace

import (
	"testing"
	"time"
)

type test struct {
	path string
	prio Priority
	tmpl string
	args []interface{}

	shouldCall  bool
	expectedMsg string
}

func xTestMT(t *testing.T) {
	testData := []test{
		test{
			path:       "trace",
			prio:       Error,
			tmpl:       "hello",
			shouldCall: true,
		},
		test{
			path:       "trace",
			prio:       Info,
			tmpl:       "hello",
			shouldCall: true,
		},
		test{
			path:       "trace",
			prio:       Debug,
			tmpl:       "hello",
			shouldCall: false,
		},

		test{
			path:       "tes",
			prio:       Error,
			tmpl:       "hello",
			shouldCall: false,
		},
		test{
			path:       "tracea",
			prio:       Error,
			tmpl:       "hello",
			shouldCall: false,
		},
		test{
			path:       "trace/a",
			prio:       Error,
			tmpl:       "hello",
			shouldCall: true,
		},

		test{
			path:        "trace",
			prio:        Error,
			tmpl:        "hello %d %d %d",
			args:        []interface{}{1, 2, 3},
			shouldCall:  true,
			expectedMsg: "hello 1 2 3",
		},
	}

	var called bool
	var seenMsg string

	handlerFn := func(t time.Time, path string, prio Priority, msg string) {
		called = true
		seenMsg = msg
	}

	listener := NewListener("test", "trace", Info, handlerFn)
	Register(listener)

	tryOne := func(idx int, run test) {
		called = false
		if run.expectedMsg == "" {
			run.expectedMsg = run.tmpl
		}

		if m, ok := M(run.path, run.prio); ok {
			T(m, run.path, run.prio, run.tmpl, run.args...)
		}

		if called != run.shouldCall {
			t.Errorf("%d: should have called listener: %v, did call: %v", idx, run.shouldCall, called)
		} else if called && seenMsg != run.expectedMsg {
			t.Errorf("expected message %q, got %q", run.expectedMsg, seenMsg)
		}
	}

	for k, run := range testData {
		tryOne(k, run)
	}

	Remove(listener)

	tryOne(-1, test{
		path:       "trace",
		prio:       Error,
		tmpl:       "hello",
		shouldCall: false,
	})
}

func TestEmptyPath(t *testing.T) {
	seen := false

	handlerFn := func(t time.Time, p string, n Priority, msg string) {
		seen = true
	}

	listener := NewListener("test", "", Info, handlerFn)
	Register(listener)
	defer Remove(listener)

	if m, ok := M("test", Info); ok {
		T(m, "test", Info, "hello")
	}

	if !seen {
		t.Error("failed to call listener")
	}
}

func handlerFunc(t time.Time, p string, n Priority, msg string) {
}

func BenchmarkFunctionCall(b *testing.B) {
	now := time.Now()
	msg := "hello"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handlerFunc(now, "test", Info, msg)
	}
}

func BenchmarkNoListeners(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if m, ok := M("/trace/a/b", Info); ok {
			T(m, "/trace/a/b", Info, "%d\n", i)
		}
	}
}

func BenchmarkOtherListeners(b *testing.B) {
	for _, path := range []string{"path1", "path2"} {
		l := NewListener(path, path, Info, handlerFunc)
		Register(l)
		defer Remove(l)
	}
	for i := 0; i < b.N; i++ {
		if m, ok := M("/elsewhere", Info); ok {
			T(m, "/elsewhere", Info, "%d\n", i)
		}
	}
}

func BenchmarkFirstListener(b *testing.B) {
	for _, path := range []string{"path1", "path2"} {
		l := NewListener(path, path, Info, handlerFunc)
		Register(l)
		defer Remove(l)
	}
	for i := 0; i < b.N; i++ {
		if m, ok := M("path1", Info); ok {
			T(m, "path1", Info, "%d\n", i)
		}
	}
}

func BenchmarkSecondListener(b *testing.B) {
	for _, path := range []string{"path1", "path2"} {
		l := NewListener(path, path, Info, handlerFunc)
		Register(l)
		defer Remove(l)
	}
	for i := 0; i < b.N; i++ {
		if m, ok := M("path2", Info); ok {
			T(m, "path2", Info, "%d\n", i)
		}
	}
}

func BenchmarkBothListeners(b *testing.B) {
	for _, path := range []string{"/trace", "/trace/a"} {
		l := NewListener(path, path, Info, handlerFunc)
		Register(l)
		defer Remove(l)
	}
	for i := 0; i < b.N; i++ {
		if m, ok := M("/trace/a/b", Info); ok {
			T(m, "/trace/a/b", Info, "%d\n", i)
		}
	}
}
