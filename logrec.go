package logrec

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	std *Logrec
)

type Level uint8

type Logrec struct {
	Out   io.Writer
	Level Level
	mutex sync.Mutex
}

const (
	LVL_DEBUG Level = iota
	LVL_INFO
	LVL_WARN
	LVL_ERROR
	LVL_FATAL
	LVL_PANIC
)

func (level Level) String() string {
	switch level {
	case LVL_DEBUG:
		return "debug"
	case LVL_INFO:
		return "info"
	case LVL_WARN:
		return "warn"
	case LVL_ERROR:
		return "error"
	case LVL_FATAL:
		return "fatal"
	case LVL_PANIC:
		return "panic"
	}

	return ""
}

func parseLevel(level string) (Level, error) {
	switch level {
	case "debug":
		return LVL_DEBUG, nil
	case "info":
		return LVL_INFO, nil
	case "warn":
		return LVL_WARN, nil
	case "error":
		return LVL_ERROR, nil
	case "fatal":
		return LVL_FATAL, nil
	case "panic":
		return LVL_PANIC, nil
	}

	return LVL_DEBUG, errors.New("Invalid level string.")
}

func New(logFile, logLevel string) error {
	level, err := parseLevel(logLevel)
	if err != nil {
		return err
	}

	fileRotator := &FileRotator{
		FileName: logFile,
	}

	std = &Logrec{
		Out:   fileRotator,
		Level: level,
	}

	return nil
}

func trace(depth int) (map[string]string, error) {
	pc, file, line, ok := runtime.Caller(depth)
	if !ok {
		return nil, errors.New("runtime caller failed.")
	}

	path := strings.Split(file, "/")
	fname := path[len(path)-1]
	funcname := runtime.FuncForPC(pc).Name()

	mFileInfo := make(map[string]string, 3)
	mFileInfo["func"] = funcname
	mFileInfo["file"] = fname
	mFileInfo["line"] = strconv.Itoa(line)

	return mFileInfo, nil
}

func (log *Logrec) styleFormat(msg string, mFileInfo map[string]string) string {
	b := &bytes.Buffer{}

	nowTime := time.Now().Format("2006-01-02 15:04:05")

	// format msg, trim the char [' ' '\n' '\r']
	msg = strings.TrimLeft(msg, "\r\n ")
	msg = strings.TrimRight(msg, "\r\n ")

	fmt.Fprintf(b, "%s %s [%s] %s [%s, %s]\n", nowTime, mFileInfo["func"], log.Level.String(),
		msg, mFileInfo["file"], mFileInfo["line"])

	return b.String()
}

func (log *Logrec) Log(lvl Level, msg string) {
	// get the line, filename, funcname of current call
	fileInfo, _ := trace(3)

	formatMsg := log.styleFormat(msg, fileInfo)

	log.mutex.Lock()
	io.WriteString(log.Out, formatMsg)
	log.mutex.Unlock()
}

func Debugf(format string, args ...interface{}) {
	if std.Level <= LVL_DEBUG {
		std.Log(LVL_DEBUG, fmt.Sprintf(format, args...))
	}
}

func Infof(format string, args ...interface{}) {
	if std.Level <= LVL_INFO {
		std.Log(LVL_INFO, fmt.Sprintf(format, args...))
	}
}

func Warnf(format string, args ...interface{}) {
	if std.Level <= LVL_WARN {
		std.Log(LVL_WARN, fmt.Sprintf(format, args...))
	}
}

func Errorf(format string, args ...interface{}) {
	if std.Level <= LVL_ERROR {
		std.Log(LVL_ERROR, fmt.Sprintf(format, args...))
	}
}

func Fatalf(format string, args ...interface{}) {
	if std.Level <= LVL_FATAL {
		std.Log(LVL_FATAL, fmt.Sprintf(format, args...))
	}
	os.Exit(1)
}

func Panicf(format string, args ...interface{}) {
	if std.Level <= LVL_PANIC {
		std.Log(LVL_PANIC, fmt.Sprintf(format, args...))
	}
}
