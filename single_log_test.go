package logrec

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNewSingleLog(t *testing.T) {
	opts := []Option{
		WithFileName("log/single"),
		WithLogLevel(LogLevelError),
	}

	logger := NewSingleLoggerWithOptions(opts...)
	assert.Equal(t, LogLevelError, logger.logLevel)

	logger.Trace("trace log")
	logger.Debug("debug log")
	logger.Error("error log")
	logger.Fatal("fatal log")

	time.Sleep(2 * time.Second)
}
