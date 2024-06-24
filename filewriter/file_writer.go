package filewriter

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gofrs/flock"
)

var (
	CheckInterval = time.Minute
)

const (
	logFileNameFormat = "%s.%4d-%02d-%02d.log"
)

// FileWriter 日志实现Writer
type FileWriter struct {
	maxSize  int64
	maxNum   int
	maxDay   int
	fileName string
	filePath string
	file     *os.File
	writer   io.Writer
	mu       sync.Mutex
	ch       chan []byte
	pending  int64      // 还有多少未写入文件的日志，用于Flush
	cond     *sync.Cond // 通知pending=0
	pmu      sync.Mutex // 配合cond使用
}

func GetFileWriter(fileName string, maxSize int64, maxNum int, maxDay int) (*FileWriter, error) {
	return defaultWriters.GetOrCreate(fileName, maxSize, maxNum, maxDay)
}

func NewFileWriter(fileName string, maxSize int64, maxNum int, maxDay int) (*FileWriter, error) {
	y, m, d := time.Now().Date()
	logPath := fmt.Sprintf(logFileNameFormat, fileName, y, m, d)
	if err := os.Mkdir(filepath.Dir(logPath), os.ModePerm); err != nil && !os.IsExist(err) {
		return nil, err
	}

	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	writer := &FileWriter{
		fileName: fileName,
		filePath: logPath,
		file:     file,
		writer:   file,
		ch:       make(chan []byte, 10000),
		maxSize:  maxSize,
		maxNum:   maxNum,
		maxDay:   maxDay,
	}

	writer.cond = sync.NewCond(&writer.pmu)
	go writer.rotate()
	go writer.flush()
	go writer.check()
	return writer, nil
}

// Write 异步channel写日志
func (w *FileWriter) Write(p []byte) (int, error) {
	w.pmu.Lock()
	w.pending++
	w.pmu.Unlock()

	buf := make([]byte, len(p))
	copy(buf, p)
	select {
	case w.ch <- buf:
		return len(buf), nil
	default:
		return 0, errors.New("chan full, drop")
	}
}

// Flush 将当前channel缓存的日志写磁盘
func (w *FileWriter) Flush() {
	w.pmu.Lock()
	for w.pending > 0 {
		w.cond.Wait()
	}
	w.pmu.Unlock()
}

// check 每分钟检查一下日志文件是否存在，运维误删log文件但是进程一直在打日志，fd会一直存在，需要关闭。超过maxSize自动rotate
func (w *FileWriter) check() {
	for {
		time.Sleep(CheckInterval)

		w.mu.Lock()
		_, err := os.Stat(w.filePath)
		if os.IsNotExist(err) || (w.file != nil && w.file.Name() == os.DevNull) {
			w.file.Close()
			file, e := os.OpenFile(w.filePath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
			if e == nil {
				w.file = file
				w.writer = file
			} else {
				w.file, _ = os.Open(os.DevNull)
				w.writer = io.Discard
			}
			w.mu.Unlock()
			continue
		}
		fileInfo, err := w.file.Stat()
		if err != nil {
			w.mu.Unlock()
			continue
		}
		if w.maxSize > 0 && fileInfo.Size() > w.maxSize {
			// protect when multiple processes writing to the same file and doing rotation
			fileLock := flock.New(path.Dir(w.filePath) + "/lock")
			_ = fileLock.Lock()

			name := path.Base(w.filePath) + ".full."
			files, _ := os.ReadDir(path.Dir(w.filePath))
			var minNum = 1000000
			var maxNum = 0
			var totalNum = 0
			for _, f := range files {
				if strings.Contains(f.Name(), name) {
					totalNum++
					s := strings.Split(f.Name(), ".")
					if len(s) > 3 {
						l := len(s) - 2
						n, _ := strconv.Atoi(s[l])
						if n > maxNum {
							maxNum = n
						}
						if n < minNum {
							minNum = n
						}
					}
				}
			}
			w.file.Close()
			//rename log file
			if pathInfo, err := os.Stat(w.filePath); err == nil && pathInfo.Size() > w.maxSize {
				name = fmt.Sprintf("%s.full.%d.log", w.filePath, maxNum+1) //织云日志清理规则 默认需要以 .log 结尾
				err := os.Rename(w.filePath, name)
				if err != nil {
					fmt.Printf("rename file path:%s fail:%s\n", w.filePath, err)
				}
			}
			file, err := os.OpenFile(w.filePath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
			if err != nil {
				fmt.Printf("open file path:%s fail:%s\n", w.filePath, err)
			} else {
				w.file = file
				w.writer = file
			}
			if totalNum >= w.maxNum {
				//remove oldest log file
				name = fmt.Sprintf("%s.full.%d.log", w.filePath, minNum)
				if err := os.Remove(name); err != nil {
					fmt.Printf("remove file path:%s fail:%s\n", name, err)
				}
			}

			_ = fileLock.Unlock()
		}
		w.mu.Unlock()
	}
}

// rotate 按天更新日志文件名
func (w *FileWriter) rotate() {
	for {
		now := time.Now()
		y, m, d := now.Add(24 * time.Hour).Date()
		nextDay := time.Date(y, m, d, 0, 0, 0, 0, now.Location())
		tm := time.NewTimer(time.Duration(nextDay.UnixNano() - now.UnixNano() - 100))
		<-tm.C
		w.mu.Lock()
		logPath := fmt.Sprintf(logFileNameFormat, w.fileName, y, m, d)

		if file, err := os.OpenFile(logPath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666); err == nil {
			w.file.Close()
			w.file = file
			w.writer = file
			w.filePath = logPath
		}

		// 删除 maxDay 天前的日志
		y, m, d = now.Add(-time.Duration(w.maxDay) * 24 * time.Hour).Date()
		files, err := filepath.Glob(fmt.Sprintf(logFileNameFormat+"*", w.fileName, y, m, d))
		if err == nil {
			for _, f := range files {
				os.Remove(f)
			}
		}
		w.mu.Unlock()
	}
}

// flush 刷新日志到磁盘中
func (w *FileWriter) flush() {
	for {
		logData := <-w.ch
		w.mu.Lock()
		w.writer.Write(logData)
		w.mu.Unlock()

		w.pmu.Lock()
		w.pending--
		if w.pending == 0 {
			w.cond.Signal()
		}
		w.pmu.Unlock()
	}
}
