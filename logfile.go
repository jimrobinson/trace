package trace

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// LogWriter implements an io.Writer that can write to a logfile,
// optionally using a useTimeFmt path name based on the time.
type LogWriter struct {
	// dir holds the directory to write the log file to
	dir string
	// name holds a log path, when useTimeFmt is true the value may contain a
	// time.Time.Format compatible sub-string
	name string
	// useTimeFmt flags whether or not path is a time-based filename.
	useTimeFmt bool
	// perm
	perm os.FileMode
	// fh is the open filehandle to the current logfile.
	fh *os.File
	// stat is the os.FileInfo of fh.Name(), when fh is open
	stat os.FileInfo
	// mu controls concurrent access to fh
	mu *sync.Mutex
	// fmtFn implements the log message formatter
	fmtFn FormatterFn
}

// NewFileLogWriter initializes a new LogWriter using an already open *os.File
// as the destination for output.
func NewFileLogWriter(fh *os.File, fmtFn FormatterFn) (w *LogWriter, err error) {
	if fh == nil {
		return nil, fmt.Errorf("NewFileLogWriter: specified filehandle is nil")
	}

	dir, name := filepath.Split(fh.Name())
	w = &LogWriter{
		dir:        dir,
		name:       name,
		useTimeFmt: false,
		fh:         fh,
		stat:       nil,
		mu:         &sync.Mutex{},
		fmtFn:      fmtFn,
	}

	return w, nil
}

// NewLogWriter initializes a new LogWriter using a file named name
// in directory dir.  The supplied perm is used to set the permissions
// of the log file when it is created.  The supplied FormatterFn will
// be used to format the messages;  adding a newline to the message
// if one is not produced by fmtFn.
func NewLogWriter(dir, name string, perm os.FileMode, fmtFn FormatterFn) (w *LogWriter, err error) {
	stat, err := os.Stat(dir)
	if err != nil {
		return nil, fmt.Errorf("unable to stat dir %s: %v", dir, err)
	}
	if !stat.IsDir() {
		return nil, fmt.Errorf("specified dir path is not a directory: %s", dir)
	}

	w = &LogWriter{
		dir:        dir,
		name:       name,
		useTimeFmt: false,
		perm:       perm,
		fh:         nil,
		stat:       nil,
		mu:         &sync.Mutex{},
		fmtFn:      fmtFn,
	}
	err = w.checkPath()
	return w, err
}

// NewTimeLogWriter initializes a new LogWriter.  The name is expected
// to contain a time.Time Format compatible sub-string, for example the literal string
// "2006-01-02.log"  will create a log filepath based on the current
// year, month, and day  The supplied perm is used to set the permissions
// of the log file when it is created.  The supplied FormatterFn will
// be used to format the messages;  adding a newline to the message
// if one is not produced by fmtFn.
func NewTimeLogWriter(dir, name string, perm os.FileMode, fmtFn FormatterFn) (w *LogWriter, err error) {
	stat, err := os.Stat(dir)
	if err != nil {
		return nil, fmt.Errorf("unable to stat dir %s: %v", dir, err)
	}
	if !stat.IsDir() {
		return nil, fmt.Errorf("specified dir path is not a directory: %s", dir)
	}

	w = &LogWriter{
		dir:        dir,
		name:       name,
		useTimeFmt: true,
		perm:       perm,
		fh:         nil,
		stat:       nil,
		mu:         &sync.Mutex{},
		fmtFn:      fmtFn,
	}
	err = w.checkPath()
	return w, err
}

// ListenerFn provides a hook to register a LogWriter with the trace
// framework.  It will use the LogWriter FormatterFn to format the
// message, adding a newline if one is not produced by the FormatterFn,
// and writing the result to the current log filepath.
func (w *LogWriter) ListenerFn(t time.Time, path string, priority Priority, entry string, args ...interface{}) {
	msg := w.fmtFn(t, path, priority, entry, args...)

	var buf []byte
	if strings.HasSuffix(msg, "\n") {
		buf = []byte(msg)
	} else {
		n := len(msg) + 1
		buf = make([]byte, n)
		copy(buf, msg)
		buf[n-1] = '\n'
	}

	w.Write(buf)
}

func (w *LogWriter) Name() string {
	w.mu.Lock()
	defer w.mu.Unlock()
	if err := w.checkPath(); err != nil {
		return ""
	}
	return w.fh.Name()
}

func (w *LogWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if err = w.checkPath(); err != nil {
		return 0, err
	}
	return w.fh.Write(p)
}

func (w *LogWriter) Close() (err error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.fh != nil {
		err = w.fh.Close()
		w.stat = nil
	}
	return err
}

// checkPath ensures w.fh is open and pointing to the correct filename
func (w *LogWriter) checkPath() error {

	var path string
	if w.useTimeFmt {
		path = filepath.Join(w.dir, time.Now().Format(w.name))
	} else {
		path = filepath.Join(w.dir, w.name)
	}

	// if the filehandle has not yet been opened, or if it is open but path
	// has changed
	var err error
	if w.fh == nil || w.fh.Name() != path {
		err = w.openFile(path)
	} else {
		// if a previously opened file has been renamed
		var stat os.FileInfo
		stat, err = os.Stat(path)
		if err != nil || !os.SameFile(w.stat, stat) {
			err = w.openFile(path)
		}
	}

	return err
}

// openFile calls os.OpenFile on path, if w.fh is not nil
//  it will be closed before being re-opened.
func (w *LogWriter) openFile(path string) error {
	var err error

	if w.fh != nil {
		w.fh.Close()
	}

	w.fh, err = os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, w.perm)
	if err != nil {
		return err
	}

	w.stat, err = os.Stat(path)

	return err
}
