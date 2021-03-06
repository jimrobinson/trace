package trace

import (
	"fmt"
	"strings"
)

type Priority uint8

const (
	Trace Priority = iota
	Debug
	Info
	Warn
	Error
	None
)

func ParsePriority(s string) (level Priority, err error) {
	level = Error
	switch strings.ToLower(s) {
	case "trace":
		level = Trace
	case "debug":
		level = Debug
	case "info":
		level = Info
	case "warn":
		level = Warn
	case "error":
		level = Error
	case "none":
		level = None
	default:
		err = fmt.Errorf("valid trace priorities are: trace, debug, info, warn, error, or none")
	}
	return level, err
}

func (p Priority) Next() Priority {
	switch p {
	case None:
		return Error
	case Error:
		return Warn
	case Warn:
		return Info
	case Info:
		return Debug
	case Debug:
		return Trace
	case Trace:
		return None
	default:
		return None
	}
}

func (p Priority) String() string {
	switch p {
	case Trace:
		return "Trace"
	case Debug:
		return "Debug"
	case Info:
		return "Info"
	case Warn:
		return "Warn"
	case Error:
		return "Error"
	case None:
		return "None"
	default:
		return fmt.Sprintf("%d", int(p))
	}
}
