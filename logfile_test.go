package trace

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewLogWriter(t *testing.T) {
	dir, err := ioutil.TempDir("", "trace_logfile.")
	if err != nil {
		t.Fatalf("unable to open tempfile: %v", err)
	}

	defer os.RemoveAll(dir)

	for i := 0; i < 5; i++ {
		name := fmt.Sprintf("%d.log", i)
		expected := filepath.Join(dir, name)

		w, err := NewLogWriter(dir, name, 0644, DefaultFormatterFn)
		if err != nil {
			t.Errorf("NewLogWriter error: %v", err)
			continue
		}

		defer w.Close()

		if expected != w.Name() {
			t.Errorf("expected logpath [%s] got [%s]", expected, w.Name())
		}
	}
}

func TestNewTimeLogWriter(t *testing.T) {
	dir, err := ioutil.TempDir("", "trace_logfile.")
	if err != nil {
		t.Fatalf("unable to open tempfile: %v", err)
	}

	defer os.RemoveAll(dir)

	name := "2006-01-02.log"
	expected := filepath.Join(dir, time.Now().Format(name))

	w, err := NewTimeLogWriter(dir, name, 0644, DefaultFormatterFn)
	if err != nil {
		t.Errorf("NewTimeLogWriter error: %v", err)
	}

	defer w.Close()

	if expected != w.Name() {
		t.Errorf("expected logpath [%s] got [%s]", expected, w.Name())
	}
}

func TestLogWriterCheckPath(t *testing.T) {
	dir, err := ioutil.TempDir("", "trace_logfile.")
	if err != nil {
		t.Fatalf("unable to open tempfile: %v", err)
	}

	defer os.RemoveAll(dir)

	timeFmt := "2006-01-02.15:04:05"
	name := fmt.Sprintf("%s.log", timeFmt)

	w, err := NewTimeLogWriter(dir, name, 0644, DefaultFormatterFn)
	if err != nil {
		t.Errorf("NewTimeLogWriter error: %v", err)
	}

	defer w.Close()

	// the following assumes enough cpu cycles are free to test the filename
	// within a millisecond of the sleep waking
	for i := 0; i < 5; i++ {
		now := time.Now()
		z := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second(), 999999, now.Location())
		z.Add(time.Millisecond)
		time.Sleep(time.Until(z))

		logpath := w.Name()
		expected := filepath.Join(dir, z.Format(name))
		if expected != logpath {
			t.Errorf("expected logpath [%s] got [%s]", expected, logpath)
			break
		}
	}
}

func TestLogWriterListenerFn(t *testing.T) {
	dir, err := ioutil.TempDir("", "trace_logfile.")
	if err != nil {
		t.Fatalf("unable to open tempfile: %v", err)
	}

	defer os.RemoveAll(dir)

	w, err := NewLogWriter(dir, "test.log", 0644, DefaultFormatterFn)
	if err != nil {
		t.Fatalf("unable to open new LogWriter: %v", err)
	}

	defer w.Close()

	var expected []string
	for i := 0; i < 10; i++ {
		tmFmt := fmt.Sprintf("2006-01-02T15:04:%02d-07:00", i)

		tm, err := time.Parse(time.RFC3339, tmFmt)
		if err != nil {
			t.Errorf("unable to parse test input time: %v", err)
		}

		w.ListenerFn(tm, "github.com/jimrobinson/trace", Trace, "%s %d %0.2f", "hello, world!", 3, 0.009)
		expected = append(expected, fmt.Sprintf("[%s][github.com/jimrobinson/trace] hello, world! 3 0.01", tmFmt))
	}

	fh, err := os.Open(w.Name())
	if err != nil {
		t.Errorf("unable to open log path: %s: %v", w.Name(), err)
	}

	defer fh.Close()

	var actual []string
	br := bufio.NewReader(fh)
	lineno := 0
	for {
		l, cont, err := br.ReadLine()
		lineno++
		if err != nil {
			if err != io.EOF {
				t.Errorf("unexpected error reading %s: %v", fh.Name(), err)
			}
			break
		}
		if cont {
			t.Errorf("unexpected continuation from bufio.Reader: %s:%d: %s", fh.Name(), lineno, string(l))
			break
		}
		actual = append(actual, string(l))
	}

	if len(expected) != len(actual) {
		t.Errorf("expected %d lines, got %d", len(expected), len(actual))
	}

	for i, v := range expected {
		if v != actual[i] {
			t.Errorf("[%d] expected line [%s] got [%s]", i, v, actual[i])
		}
	}
}

func BenchmarkTimeLogWriterListenerFn(b *testing.B) {
	name := "2006-01-02.log"

	dir, err := ioutil.TempDir("", "trace_logfile.")
	if err != nil {
		b.Fatalf("unable to open tempfile: %v", err)
	}

	defer os.RemoveAll(dir)

	w, err := NewTimeLogWriter(dir, name, 0644, DefaultFormatterFn)
	if err != nil {
		b.Errorf("NewTimeLogWriter error: %v", err)
	}

	defer w.Close()

	dt := time.Now()
	path := "github.com/highwire/jimr/trace"
	priority := []Priority{
		Trace,
		Debug,
		Info,
		Warn,
		Error,
	}
	format := "BenchMark iteration %d of %d"

	b.StartTimer()
	for n := 0; n < b.N; n++ {
		w.ListenerFn(dt, path, priority[n%len(priority)], format, n, b.N)
		dt.Add(time.Nanosecond)
	}
	b.StopTimer()
}
