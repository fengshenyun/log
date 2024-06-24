package logrec

import (
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/fengshenyun/logrec/filewriter"
	"github.com/goccy/go-json"
	"github.com/valyala/bytebufferpool"
)

var (
	bufferPool bytebufferpool.Pool // 字节内存缓存池，减少内存分配和回收
)

type BulkLogger struct {
	LogLevel int
	flag     int
	writer   io.Writer
	mu       sync.Mutex
	maxLevel int
	buf      *bytebufferpool.ByteBuffer
	ext      map[string]interface{}
	logTime  time.Time
}

func NewBulkLoggerWithOptions(opts ...Option) *BulkLogger {
	logConf := NewLogConf()
	for _, opt := range opts {
		opt(logConf)
	}

	writer, err := filewriter.GetFileWriter(logConf.fileName, logConf.maxSize, logConf.maxNum, logConf.maxDay)
	if err != nil {
		panic(err)
	}

	return NewBulkLogger(logConf.level, writer, logConf.flag)
}

func NewBulkLogger(logLevel int, writer io.Writer, flag int) *BulkLogger {
	logger := &BulkLogger{
		buf: bufferPool.Get(),
		ext: make(map[string]interface{}),
	}

	logger.LogLevel = logLevel
	logger.writer = writer
	logger.flag = flag
	return logger
}

func (l *BulkLogger) Level() int {
	return l.maxLevel
}

func (l *BulkLogger) SetMaxLevel(logLevel int) {
	l.mu.Lock()
	if logLevel > l.maxLevel {
		l.maxLevel = logLevel
	}
	l.mu.Unlock()
}

func (l *BulkLogger) Trace(format string, v ...interface{}) {
	l.Writew(LogLevelTrace, format, v...)
}

func (l *BulkLogger) Debug(format string, v ...interface{}) {
	l.Writew(LogLevelDebug, format, v...)
}

func (l *BulkLogger) Info(format string, v ...interface{}) {
	l.Writew(LogLevelInfo, format, v...)
}

func (l *BulkLogger) Warning(format string, v ...interface{}) {
	l.Writew(LogLevelWarning, format, v...)
}

func (l *BulkLogger) Error(format string, v ...interface{}) {
	l.Writew(LogLevelError, format, v...)
}

func (l *BulkLogger) Fatal(format string, v ...interface{}) {
	l.Writew(LogLevelFatal, format, v...)
}

func (l *BulkLogger) Writew(level int, template string, args ...interface{}) {
	if l == nil {
		return
	}

	now := time.Now()
	_, file, line, _ := Caller(3)
	var fieldIdx []int
	var extFields []Field
	for i := range args {
		switch v := args[i].(type) {
		case Field:
			extFields = append(extFields, v)
			fieldIdx = append(fieldIdx, i)
		}
	}

	var fmtArgs []interface{}
	if len(fieldIdx) == 0 {
		fmtArgs = args
	} else if len(fieldIdx) < len(args) {
		fmtArgs = make([]interface{}, 0, len(args)-len(fieldIdx))
		idx := 0
		for i := range args {
			if idx < len(fieldIdx) && fieldIdx[idx] == i {
				idx++
				continue
			}
			fmtArgs = append(fmtArgs, args[i])
		}
	}
	if len(fmtArgs) > 0 {
		template = fmt.Sprintf(template, fmtArgs...)
	}

	l.wLock()
	defer l.wUnlock()

	l.logTime = now
	if l.maxLevel < level {
		l.maxLevel = level
	}
	l.formatHeader(now, file, line)
	l.buf.WriteByte('[')
	l.buf.WriteString(LogLevel(level).String())
	l.buf.WriteString("] ")
	l.buf.WriteString(template)
	l.buf.WriteByte('\n')

	for _, field := range extFields {
		l.ext[field.Key] = field.Value
	}
}

// Finishw 结束添加日志，添加一行结构化的包含通用字段(logging.Fields)，
// 以及用户自定义的结构化字段的json后，由writer控制写入文件
func (l *BulkLogger) Finishw() {
	if l == nil {
		return
	}

	l.wLock()
	defer func() {
		l.reset()
		l.wUnlock()
	}()

	if l.buf.Len() == 0 {
		return
	}

	if l.maxLevel < l.LogLevel {
		return
	}

	// 打印通用字段
	if len(l.ext) > 0 {
		extData, extErr := json.Marshal(l.ext)
		if extErr != nil {
			return
		}

		l.buf.Write(extData)
		l.buf.WriteByte('\n')
	}

	l.buf.WriteString("==>\n")
	l.writer.Write(l.buf.Bytes())
}

func (l *BulkLogger) wLock() {
	l.mu.Lock()
	if l.buf == nil {
		l.buf = bufferPool.Get()
	}

	if l.ext == nil {
		l.ext = make(map[string]interface{})
	}
}

func (l *BulkLogger) wUnlock() {
	l.mu.Unlock()
}

func (l *BulkLogger) reset() {
	l.maxLevel = LogLevelNull
	l.logTime = time.Time{}
	if l.buf != nil {
		bufferPool.Put(l.buf)
		l.buf = nil
	}
	l.ext = nil
}

func (l *BulkLogger) formatHeader(t time.Time, file string, line int) {
	if l.flag&(log.Ldate|log.Ltime|log.Lmicroseconds) != 0 {
		if l.flag&log.LUTC != 0 {
			t = t.UTC()
		}
		if l.flag&log.Ldate != 0 {
			year, month, day := t.Date()
			itoa(l.buf, year, 4)
			l.buf.WriteByte('/')
			itoa(l.buf, int(month), 2)
			l.buf.WriteByte('/')
			itoa(l.buf, day, 2)
			l.buf.WriteByte(' ')
		}
		if l.flag&(log.Ltime|log.Lmicroseconds) != 0 {
			hour, min, sec := t.Clock()
			itoa(l.buf, hour, 2)
			l.buf.WriteByte(':')
			itoa(l.buf, min, 2)
			l.buf.WriteByte(':')
			itoa(l.buf, sec, 2)
			if l.flag&log.Lmicroseconds != 0 {
				l.buf.WriteByte('.')
				itoa(l.buf, t.Nanosecond()/1e3, 6)
			} else {
				l.buf.WriteByte('.')
				itoa(l.buf, t.Nanosecond()/1e6, 3)
			}
			l.buf.WriteByte(' ')
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
		l.buf.WriteString(file)
		l.buf.WriteByte(':')
		itoa(l.buf, line, -1)
		l.buf.WriteString(": ")
	}
}

func itoa(buf io.Writer, i int, wid int) {
	// Assemble decimal in reverse order.
	var b [20]byte
	bp := len(b) - 1
	for i >= 10 || wid > 1 {
		wid--
		q := i / 10
		b[bp] = byte('0' + i - q*10)
		bp--
		i = q
	}
	// i < 10
	b[bp] = byte('0' + i)
	buf.Write(b[bp:])
}
