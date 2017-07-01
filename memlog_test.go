package trace

import (
	"bufio"
	"container/list"
	"fmt"
	"io"
	"reflect"
	"sync"
	"testing"
	"time"
)

var eventsReaderEntries = [][]string{
	[]string{},
	[]string{
		"a",
	},
	[]string{
		"b",
		"c",
	},
	[]string{
		"d",
		"e",
		"f",
	},
}

func TestEventsReader(t *testing.T) {
	for i := 0; i < len(eventsReaderEntries); i++ {
		eventSet := eventsReaderEntries[i]

		events := list.New()
		for j := 0; j < len(eventSet); j++ {
			events.PushFront(eventSet[j])
		}

		snapshot := make([]*list.Element, 0, events.Len())
		event := events.Front()
		for event != nil {
			snapshot = append(snapshot, event)
			event = event.Next()
		}

		r := newEventsReader(snapshot)
		br := bufio.NewReader(r)

		expected := make([]string, len(eventSet))
		for j, k := len(eventSet)-1, 0; j >= 0; j, k = j-1, k+1 {
			expected[k] = eventSet[j]
		}

		for {
			msg, cont, err := br.ReadLine()
			if err != nil {
				if err != io.EOF {
					t.Error(err)
				}
				if len(expected) != 0 {
					t.Errorf("expected [%s] got io.EOF", expected[0])
				}
				break
			}
			if cont {
				t.Error("unexpected state: line length should not have been exceeded")
				break
			}
			if len(expected) == 0 {
				t.Errorf("expected io.EOF, got [%s]", string(msg))
				break
			}
			if expected[0] != string(msg) {
				t.Errorf("expected [%s] got [%s]", expected[0], string(msg))
			}

			expected = expected[1:]
		}
	}
}

func TestPriorityLogReader(t *testing.T) {

	for i := 0; i < len(eventsReaderEntries); i++ {
		eventSet := eventsReaderEntries[i]

		events := list.New()
		bytes := 0
		for j := 0; j < len(eventSet); j++ {
			bytes += len(eventSet[j])
			events.PushFront(eventSet[j])
		}

		plog := &priorityLog{
			limitEntries: len(eventSet),
			limitBytes:   bytes,
			messages:     events,
			mu:           &sync.RWMutex{},
		}

		// test Reader w/o line limit
		r := plog.Reader(-1)
		br := bufio.NewReader(r)

		expected := make([]string, len(eventSet))
		for j, k := len(eventSet)-1, 0; j >= 0; j, k = j-1, k+1 {
			expected[k] = eventSet[j]
		}

		for {
			msg, cont, err := br.ReadLine()
			if err != nil {
				if err != io.EOF {
					t.Error(err)
				}
				if len(expected) != 0 {
					t.Errorf("expected [%s] got io.EOF", expected[0])
				}
				break
			}
			if cont {
				t.Error("unexpected state: line length should not have been exceeded")
				break
			}
			if len(expected) == 0 {
				t.Errorf("expected io.EOF, got [%s]", string(msg))
				break
			}
			if expected[0] != string(msg) {
				t.Errorf("expected [%s] got [%s]", expected[0], string(msg))
			}

			expected = expected[1:]
		}

		// test Reader w/ line limit 1
		r = plog.Reader(1)
		br = bufio.NewReader(r)

		expected = make([]string, 0)
		for j, k := len(eventSet)-1, 0; j >= 0; j, k = j-1, k+1 {
			expected = append(expected, eventSet[j])
			break
		}

		for {
			msg, cont, err := br.ReadLine()
			if err != nil {
				if err != io.EOF {
					t.Error(err)
				}
				if len(expected) != 0 {
					t.Errorf("expected [%s] got io.EOF", expected[0])
				}
				break
			}
			if cont {
				t.Error("unexpected state: line length should not have been exceeded")
				break
			}
			if len(expected) == 0 {
				t.Errorf("expected io.EOF, got [%s]", string(msg))
				break
			}
			if expected[0] != string(msg) {
				t.Errorf("expected [%s] got [%s]", expected[0], string(msg))
			}

			expected = expected[1:]
		}
	}
}

type pushTest struct {
	limitEntries int
	limitBytes   int
	Events       []string
	Expect       []string
	err          error
}

var pushTests = []pushTest{
	{
		limitEntries: 0,
		err:          LimitEntriesErr,
	},
	{
		limitEntries: 1,
		limitBytes:   0,
		err:          LimitBytesErr,
	},
	{
		limitEntries: 1,
		limitBytes:   10,
		Events: []string{
			"0000000000",
			"1111111111",
			"2222222222",
		},
		Expect: []string{
			"2222222222",
		},
	},
	{
		limitEntries: 1000,
		limitBytes:   5,
		Events: []string{
			"a",
			"b",
			"c",
			"d",
			"e",
			"f",
			"g",
		},
		Expect: []string{
			"g",
			"f",
			"e",
			"d",
			"c",
		},
	},
}

func TestPriorityLogPush(t *testing.T) {
	for i, v := range pushTests {
		plog, err := newPriorityLog(v.limitEntries, v.limitBytes)
		if err != nil {
			if !reflect.DeepEqual(err, v.err) {
				t.Errorf("[%d] unexpected error: %v", i, err)
			}
			continue
		} else {
			if v.err != nil {
				t.Errorf("[%d] did not recieve expected error: %v", i, v.err)
			}
		}

		for _, msg := range v.Events {
			plog.push(msg)
		}

		if plog.messages.Len() != len(v.Expect) {
			t.Errorf("expected %d messages in plog, got %d", len(v.Expect), plog.messages.Len())
		}

		e := plog.messages.Front()
		for j, s := range v.Expect {
			if e == nil {
				t.Errorf("[%d/%d] expected [%s] but got nil", i, j, s)
				continue
			}
			if e.Value.(string) != s {
				t.Errorf("[%d/%d] expected [%s] but got [%s]", i, j, s, e.Value.(string))
				continue
			}
			e = e.Next()
		}
		if e != nil {
			t.Errorf("[%d] expected nil but got [%s]", i, e.Value.(string))
		}
	}
}

type memLogTestSend struct {
	t        time.Time
	path     string
	priority Priority
	format   string
	args     []interface{}
}

type memLogTest struct {
	limits  MemLogLimits
	backlog int
	fmtFn   FormatterFn
	err     error
	send    []memLogTestSend
	expect  map[Priority][]string
}

var memLogTests = []memLogTest{
	{
		limits:  MemLogLimits{},
		backlog: 0,
		fmtFn:   DefaultFormatterFn,
		err:     fmt.Errorf("limits must contain at least one entry"),
		send:    []memLogTestSend{},
		expect:  map[Priority][]string{},
	},
	{
		limits:  DefaultMemLogLimits,
		backlog: 0,
		fmtFn:   DefaultFormatterFn,
		err:     nil,
		send:    []memLogTestSend{},
		expect:  map[Priority][]string{},
	},
	{
		limits:  DefaultMemLogLimits,
		backlog: 100,
		fmtFn:   DefaultFormatterFn,
		err:     nil,
		send: []memLogTestSend{
			{
				t:        time.Date(2017, 06, 01, 12, 13, 14, 7, time.UTC),
				path:     "github.com/jimrobinson/trace",
				priority: Trace,
				format:   "%s",
				args:     []interface{}{"hello, world!"},
			},
		},
		expect: map[Priority][]string{
			Trace: []string{
				"[2017-06-01T12:13:14Z][github.com/jimrobinson/trace] hello, world!",
			},
		},
	},
	{
		limits:  DefaultMemLogLimits,
		backlog: 100,
		fmtFn:   DefaultFormatterFn,
		err:     nil,
		send: []memLogTestSend{
			{
				t:        time.Date(2017, 06, 01, 12, 13, 14, 7, time.UTC),
				path:     "github.com/jimrobinson/trace",
				priority: Trace,
				format:   "%s",
				args:     []interface{}{"a: hello, world!"},
			},
			{
				t:        time.Date(2017, 06, 01, 12, 13, 15, 8, time.UTC),
				path:     "github.com/jimrobinson/trace",
				priority: Trace,
				format:   "%s",
				args:     []interface{}{"b: hello, world!"},
			},
			{
				t:        time.Date(2017, 06, 01, 12, 13, 16, 9, time.UTC),
				path:     "github.com/jimrobinson/trace",
				priority: Trace,
				format:   "%s",
				args:     []interface{}{"c: hello, world!"},
			},
			{
				t:        time.Date(2017, 06, 01, 12, 13, 17, 10, time.UTC),
				path:     "github.com/jimrobinson/trace",
				priority: Trace,
				format:   "%s",
				args:     []interface{}{"d: hello, world!"},
			},
			{
				t:        time.Date(2017, 06, 01, 12, 13, 18, 10, time.UTC),
				path:     "github.com/jimrobinson/trace",
				priority: Trace,
				format:   "%s",
				args:     []interface{}{"e: hello, world!"},
			},
		},
		expect: map[Priority][]string{
			Trace: []string{
				"[2017-06-01T12:13:18Z][github.com/jimrobinson/trace] e: hello, world!",
				"[2017-06-01T12:13:17Z][github.com/jimrobinson/trace] d: hello, world!",
				"[2017-06-01T12:13:16Z][github.com/jimrobinson/trace] c: hello, world!",
				"[2017-06-01T12:13:15Z][github.com/jimrobinson/trace] b: hello, world!",
				"[2017-06-01T12:13:14Z][github.com/jimrobinson/trace] a: hello, world!",
			},
		},
	},
}

func TestMemLogListenerFn(t *testing.T) {

	for i, v := range memLogTests {
		mlog, err := NewMemLog(v.limits, v.backlog, v.fmtFn)
		if err != nil {
			if !reflect.DeepEqual(err, v.err) {
				t.Errorf("[%d] unexpected error: %v", i, err)
			}
			continue
		} else {
			if v.err != nil {
				t.Errorf("[%d] did not recieve expected error: %v", i, v.err)
			}
		}

		for _, event := range v.send {
			mlog.ListenerFn(event.t, event.path, event.priority, event.format, event.args...)
		}

		mlog.wg.Wait()

		for priority, plog := range mlog.messages {
			messages := plog.messages
			n := messages.Len()
			if len(v.expect[priority]) != n {
				t.Errorf("expected %d messages at priority %d, got %d", len(v.expect[priority]), priority, n)
				continue
			}
		}

		for priority, expect := range v.expect {
			plog := mlog.messages[priority]
			messages := plog.messages

			n := messages.Len()
			if len(expect) != n {
				t.Errorf("expected %d messages at priority %d, got %d", len(expect), priority, n)
				continue
			}

			e := messages.Front()
			for j := 0; j < len(expect); j++ {
				if e == nil {
					t.Errorf("[%d/%d] expected %s message [%s] got nil", i, j, priority, expect[j])
					break
				}
				if expect[j] != e.Value.(string) {
					t.Errorf("[%d/%d] expected %s message [%s] got [%s]", i, j, priority, expect[j], e.Value.(string))
				}
				e = e.Next()
			}
		}
	}
}

func BenchmarkMemLogListenerFn(b *testing.B) {
	mlog, err := NewMemLog(DefaultMemLogLimits, 1000, DefaultFormatterFn)
	if err != nil {
		b.Error(err)
	}

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
		mlog.ListenerFn(dt, path, priority[n%len(priority)], format, n, b.N)
		dt.Add(time.Nanosecond)
	}
	b.StopTimer()
}
