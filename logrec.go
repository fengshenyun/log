package logrec

import (
	"io"
	"strings"
	"sync"
)

var (
	std *Logrec
)

type LogLevel int

const (
	LogLevelNull = iota
	LogLevelTrace
	LogLevelDebug
	LogLevelInfo
	LogLevelWarning
	LogLevelError
	LogLevelFatal
)

type Logrec struct {
	Out   io.Writer
	Level LogLevel
	mutex sync.Mutex
}

func (level LogLevel) String() string {
	switch level {
	case LogLevelFatal:
		return "FATAL"
	case LogLevelError:
		return "ERROR"
	case LogLevelWarning:
		return "WARN"
	case LogLevelInfo:
		return "INFO"
	case LogLevelDebug:
		return "DEBUG"
	case LogLevelTrace:
		return "TRACE"
	default:
		return ""
	}
}

func parseLevel(level string) LogLevel {
	level = strings.ToLower(level)

	switch level {
	case "trace":
		return LogLevelTrace
	case "debug":
		return LogLevelDebug
	case "info":
		return LogLevelInfo
	case "warn":
		return LogLevelWarning
	case "error":
		return LogLevelError
	case "fatal":
		return LogLevelFatal
	}

	return LogLevelNull
}
