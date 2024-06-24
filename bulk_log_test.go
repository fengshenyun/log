package logrec

import "testing"

func TestNewBulkLoggerWithOptions(t *testing.T) {
	logger := NewBulkLoggerWithOptions()

	logger.Debug("debug log")
	logger.Trace("trace log")
	logger.Error("error log")
	logger.Fatal("fatal log")
	logger.Finishw()
}

func TestNewBulkLoggerWithOptions01(t *testing.T) {
	opts := []Option{
		WithFileName("log/bulk_log"),
		WithLogLevel(LogLevelError),
	}

	logger := NewBulkLoggerWithOptions(opts...)
	logger.Debug("debug log")
	logger.Trace("trace log")
	logger.Error("error log")
	logger.Finishw()
}
