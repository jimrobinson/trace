package trace

import "fmt"

type Priority uint8

const (
	Trace Priority = iota
	Debug
	Info
	Warn
	Error
)

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
