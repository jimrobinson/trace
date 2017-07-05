package trace

import (
	"bytes"
	"container/list"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"
)

// MemLogOrder controls the order in which the MemLog Reader returns its log events
type MemLogReaderOrder uint8

const (
	// DESC returns log events in descending order (newest to oldest)
	DESC MemLogReaderOrder = iota
	// ASC returns log events ascending order (oldest to newest)
	ASC
)

var LimitEntriesErr = fmt.Errorf("MemLogLimit.Entries must be >= 1")
var LimitBytesErr = fmt.Errorf("MemLogLimit.Bytes must be >= 1")

// MemLogLimit defines the limits to place on MemLog messages,
// restricting
// the number entries to keep and the
//  maximum allowable size.
type MemLogLimit struct {
	// Total number of log event entries to track
	Entries int
	// Maxmium number of bytes for all the log entries, if this number is
	// exceeded then older log entries will be discarded.
	Bytes int
}

// DefaultMemLogLimit defines a MemLogLimit of 1,000 entries and a
// 1-megabyte limit on the total size of the log entries.
var DefaultMemLogLimit = MemLogLimit{
	Entries: 1000,
	Bytes:   1048576,
}

// MemLogLimits maps Priority levels to MemLogLimit
type MemLogLimits map[Priority]MemLogLimit

// DefaultMemLogLimits defines a MemLogLimits that uses DefaultMemLogLimit
// for every Priority
var DefaultMemLogLimits = map[Priority]MemLogLimit{
	Trace: DefaultMemLogLimit,
	Debug: DefaultMemLogLimit,
	Info:  DefaultMemLogLimit,
	Warn:  DefaultMemLogLimit,
	Error: DefaultMemLogLimit,
}

// logEvent captures a log message and its priority level.
type logEvent struct {
	priority Priority
	msg      string
}

// MemLog implements an in-memory list of recent log entries, partiioned
// by Priority
type MemLog struct {
	limits   map[Priority]MemLogLimit
	messages map[Priority]*priorityLog
	queue    chan logEvent
	fmtFn    FormatterFn
	wg       *sync.WaitGroup
}

// NewMemLog initializes a new MemLog, using the specified limits and
// queue backlog.  If a Priority is not defined in limits, log entries
// matching that priority will be discarded.  The backlog sets the
// size of the queue buffer, allowing up to that many log entries to
// accumulate, pending their addition to the MemLog.  If this buffer
// is filled then new log messages will be discarded until the backlog
// is cleared.
func NewMemLog(limits MemLogLimits, backlog int, fmtFn FormatterFn) (*MemLog, error) {
	if len(limits) == 0 {
		return nil, fmt.Errorf("limits must contain at least one entry")
	}

	if fmtFn == nil {
		fmtFn = DefaultFormatterFn
	}

	mlog := &MemLog{
		limits:   limits,
		messages: make(map[Priority]*priorityLog, len(limits)),
		queue:    make(chan logEvent, 1+backlog),
		fmtFn:    fmtFn,
		wg:       &sync.WaitGroup{},
	}

	for priority, limit := range mlog.limits {
		plog, err := newPriorityLog(limit.Entries, limit.Bytes)
		if err != nil {
			return nil, fmt.Errorf("unable to initialize a priority log for priority level %d: %s", priority, err)
		}
		mlog.messages[priority] = plog
	}

	go mlog.run()

	return mlog, nil
}

// MemLog shuts down the MemLog.  The instance should not be used
// after Close is called.
func (mlog *MemLog) Close() {
	close(mlog.queue)
	mlog.wg.Wait()
}

// ListenerFn is used to register the MemLog with the trace framework.
func (mlog *MemLog) ListenerFn(t time.Time, path string, priority Priority, format string, args ...interface{}) {
	mlog.wg.Add(1)
	msg := mlog.fmtFn(t, path, priority, format, args...)
	select {
	case mlog.queue <- logEvent{priority: priority, msg: msg}:
	default:
		mlog.wg.Done()
	}
}

// run reads log messages from queue, adding them to the appropriate
// priorityLog.  If limits have not been specified for a Priority,
// the nessage will be discarded.
func (mlog *MemLog) run() {
	for v := range mlog.queue {
		if plog, ok := mlog.messages[v.priority]; ok {
			plog.push(v.msg)
		}
		mlog.wg.Done()
	}
}

// Reader returns an io.Reader for log messages at the specified
// priority level.  If the specified priority was not defined in the
// MemLog limits, a nil Reader will be returned.  If lines is > 0
// then the Reader will only return up to that many lines.
func (mlog *MemLog) Reader(priority Priority, lines int, order MemLogReaderOrder) io.Reader {
	if plog, ok := mlog.messages[priority]; ok {
		return plog.Reader(lines, order)
	}
	return nil
}

// priorityLog tracks the log messages for a Priority level.  Limtis
// on the number of entries to keep, and the maximum size of all the
// entries (regardless of count), are defined to keep the size within
// reasonable bounds.
type priorityLog struct {
	messages     *list.List
	limitEntries int
	limitBytes   int
	bytes        int
	mu           *sync.RWMutex
}

// newPriorityLog initializes a new priorityLog, placing a limit of
// limitEntries entries, or limitBytes space used across all log messages.
func newPriorityLog(limitEntries, limitBytes int) (*priorityLog, error) {
	if limitEntries < 1 {
		return nil, LimitEntriesErr
	}
	if limitBytes < 1 {
		return nil, LimitBytesErr
	}
	p := &priorityLog{
		limitEntries: limitEntries,
		limitBytes:   limitBytes,
		messages:     list.New(),
		mu:           &sync.RWMutex{},
	}
	return p, nil
}

// push adds msg to the priorityLog messages, discarding older log
// messages as necessary to enforce the limitEntries and limitBytes limits.
func (p *priorityLog) push(msg string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// remove older entries if we've reached maxlen
	elide := (p.messages.Len() - p.limitEntries + 1)
	for i := 0; i < elide; i++ {
		if e := p.messages.Back(); e != nil {
			p.bytes -= len(e.Value.(string))
			p.messages.Remove(e)
		}
	}

	// remove older entries if we've reached limitBytes
	if (p.bytes + len(msg)) > p.limitBytes {
		for {
			e := p.messages.Back()
			if e == nil {
				break
			}

			p.bytes -= len(e.Value.(string))
			p.messages.Remove(e)

			if (p.bytes + len(msg)) <= p.limitBytes {
				break
			}
		}
	}

	// push the log entry onto the head
	p.bytes += len(msg)
	p.messages.PushFront(msg)
}

// reader returns an io.Reader that contains the log entries in
// descending order by time  If lines is > 0 then the Reader will
// only return up to that many lines.
func (p *priorityLog) Reader(lines int, order MemLogReaderOrder) io.Reader {
	p.mu.RLock()
	snapshot := make([]*list.Element, 0, p.messages.Len())
	e := p.messages.Front()
	for e != nil {
		snapshot = append(snapshot, e)
		if len(snapshot) == lines {
			break
		}
		e = e.Next()
	}
	p.mu.RUnlock()

	if order == ASC && len(snapshot) > 1 {
		for i, j := 0, len(snapshot)-1; i < j; i, j = i+1, j-1 {
			snapshot[i], snapshot[j] = snapshot[j], snapshot[i]
		}
	}

	return newEventsReader(snapshot)
}

// eventsReader implements io.Reader for log messages, adding a newline
// to separate each log event if one is not already present.
type eventsReader struct {
	messages []*list.Element
	buf      *bytes.Buffer
}

func newEventsReader(messages []*list.Element) io.Reader {
	return &eventsReader{
		messages: messages,
		buf:      &bytes.Buffer{},
	}
}

// Read fills p with bytes from the log message buffer, returning the
// number of bytes written, or any error encountered.  If the error
// is io.EOF then there are no more log messages to read.
func (r *eventsReader) Read(p []byte) (int, error) {
	if r.buf.Len() == 0 {
		if len(r.messages) == 0 {
			return 0, io.EOF
		}
		r.buf.Reset()
		s := r.messages[0].Value.(string)
		r.buf.WriteString(s)
		if !strings.HasSuffix(s, "\n") {
			r.buf.WriteByte('\n')
		}
		r.messages = r.messages[1:]
	}
	return r.buf.Read(p)
}
