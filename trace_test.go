package trace

import (
	"testing"
	"time"
)

func handlerFunc(t time.Time, format string, a ...interface{}) {
}

func BenchmarkFunctionCall(b *testing.B) {
	now := time.Now()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handlerFunc(now, "%d\n", i)
	}
}

func BenchmarkNoListeners(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if m, ok := M("/trace/a/b", Info); ok {
			T(m, "%d\n", i)
		}
	}
}

func BenchmarkOtherListeners(b *testing.B) {
	for _, id := range []string{"path1", "path2"} {
		l := NewListener(handlerFunc, id, Info)
		Add(l)
		defer Remove(l)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if m, ok := M("/elsewhere", Info); ok {
			T(m, "%d\n", i)
		}
	}
}

func BenchmarkFirstListener(b *testing.B) {
	for _, id := range []string{"path1", "path2"} {
		l := NewListener(handlerFunc, id, Info)
		Add(l)
		defer Remove(l)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if m, ok := M("path1", Info); ok {
			T(m, "%d\n", i)
		}
	}
}

func BenchmarkSecondListener(b *testing.B) {
	for _, id := range []string{"path1", "path2"} {
		l := NewListener(handlerFunc, id, Info)
		Add(l)
		defer Remove(l)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if m, ok := M("path2", Info); ok {
			T(m, "%d\n", i)
		}
	}
}

func BenchmarkBothListeners(b *testing.B) {
	for _, id := range []string{"/trace", "/trace/a"} {
		l := NewListener(handlerFunc, id, Info)
		Add(l)
		defer Remove(l)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if m, ok := M("/trace/a/b", Info); ok {
			T(m, "%d\n", i)
		}
	}
}
