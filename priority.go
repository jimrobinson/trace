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
	default:
		err = fmt.Errorf("valid trace priorities are: trace, debug, info, warn, or error")
	}
	return level, err
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
	default:
		return fmt.Sprintf("%d", int(p))
	}
}
