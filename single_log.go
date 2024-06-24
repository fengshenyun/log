package logrec

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/fengshenyun/logrec/filewriter"
)

type SingleLogger struct {
	logLevel int
	flag     int
	writer   io.Writer
	mu       sync.Mutex
}

func NewSingleLoggerWithOptions(opts ...Option) *SingleLogger {
	logConf := NewLogConf()
	for _, opt := range opts {
		opt(logConf)
	}

	writer, err := filewriter.GetFileWriter(logConf.fileName, logConf.maxSize, logConf.maxNum, logConf.maxDay)
	if err != nil {
		panic(err)
	}

	return NewSingleLogger(logConf.level, writer, logConf.flag)
}

func NewSingleLogger(logLevel int, writer io.Writer, flag int) *SingleLogger {
	return &SingleLogger{
		logLevel: logLevel,
		writer:   writer,
		flag:     flag,
	}
}

func (l *SingleLogger) Trace(format string, v ...interface{}) {
	l.Writew(LogLevelTrace, format, v...)
}

func (l *SingleLogger) Debug(format string, v ...interface{}) {
	l.Writew(LogLevelDebug, format, v...)
}

func (l *SingleLogger) Info(format string, v ...interface{}) {
	l.Writew(LogLevelInfo, format, v...)
}

func (l *SingleLogger) Warning(format string, v ...interface{}) {
	l.Writew(LogLevelWarning, format, v...)
}

func (l *SingleLogger) Error(format string, v ...interface{}) {
	l.Writew(LogLevelError, format, v...)
}

func (l *SingleLogger) Fatal(format string, v ...interface{}) {
	l.Writew(LogLevelFatal, format, v...)
}

func (l *SingleLogger) Writew(level int, template string, args ...interface{}) {
	if l == nil || level < l.logLevel {
		return
	}

	now := time.Now()
	_, file, line, _ := Caller(3)

	if len(args) > 0 {
		template = fmt.Sprintf(template, args...)
	}

	buf := &bytes.Buffer{}
	l.formatHeader(buf, now, file, line)
	buf.WriteByte('[')
	buf.WriteString(LogLevel(level).String())
	buf.WriteString("] ")
	buf.WriteString(template)
	buf.WriteByte('\n')

	l.writer.Write(buf.Bytes())
}

func (l *SingleLogger) formatHeader(buf *bytes.Buffer, t time.Time, file string, line int) {
	if l.flag&(log.Ldate|log.Ltime|log.Lmicroseconds) != 0 {
		if l.flag&log.LUTC != 0 {
			t = t.UTC()
		}

		if l.flag&log.Ldate != 0 {
			year, month, day := t.Date()
			itoa(buf, year, 4)
			buf.WriteByte('/')
			itoa(buf, int(month), 2)
			buf.WriteByte('/')
			itoa(buf, day, 2)
			buf.WriteByte(' ')
		}

		if l.flag&(log.Ltime|log.Lmicroseconds) != 0 {
			hour, min, sec := t.Clock()
			itoa(buf, hour, 2)
			buf.WriteByte(':')
			itoa(buf, min, 2)
			buf.WriteByte(':')
			itoa(buf, sec, 2)
			if l.flag&log.Lmicroseconds != 0 {
				buf.WriteByte('.')
				itoa(buf, t.Nanosecond()/1e3, 6)
			} else {
				buf.WriteByte('.')
				itoa(buf, t.Nanosecond()/1e6, 3)
			}
			buf.WriteByte(' ')
		}
	}
	if l.flag&(log.Lshortfile|log.Llongfile) != 0 {
		if l.flag&log.Lshortfile != 0 {
			short := file
			for i := len(file) - 1; i > 0; i-- {
				if file[i] == '/' {
					short = file[i+1:]
					break
				}
			}
			file = short
		}
		buf.WriteString(file)
		buf.WriteByte(':')
		itoa(buf, line, -1)
		buf.WriteString(": ")
	}
}
