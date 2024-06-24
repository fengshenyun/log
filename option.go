package logrec

import "log"

var (
	LogFlag = log.LstdFlags | log.Lshortfile // 默认一行日志的前缀

)

type Option func(*LogConf)

type LogConf struct {
	level    int
	flag     int
	omitLvl  bool
	fileName string
	maxSize  int64
	maxNum   int
	maxDay   int
}

func NewLogConf() *LogConf {
	return &LogConf{
		level:    LogLevelDebug,
		flag:     LogFlag,
		fileName: "log/server",
		maxSize:  100 * 1024 * 1024,
		maxNum:   1,
		maxDay:   1,
	}
}

func WithLogLevel(level int) Option {
	return func(conf *LogConf) {
		conf.level = level
	}
}

func WithLogFlag(flag int) Option {
	return func(conf *LogConf) {
		conf.flag = flag
	}
}

func WithFileName(fileName string) Option {
	return func(conf *LogConf) {
		conf.fileName = fileName
	}
}

func WithMaxSize(maxSize int64) Option {
	return func(conf *LogConf) {
		conf.maxSize = maxSize
	}
}

func WithMaxNum(maxNum int) Option {
	return func(conf *LogConf) {
		conf.maxNum = maxNum
	}
}

func WithMaxDay(maxDay int) Option {
	return func(conf *LogConf) {
		conf.maxDay = maxDay
	}
}
