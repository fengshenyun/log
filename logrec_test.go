package logrec

import (
	"testing"
)

func TestNew1(t *testing.T) {
	New("debug.log", "debug")

	Debugf("hello world")
	Debugf("you name:%s", "duyayun")
	Warnf("warn")
	Infof("info")
}

func TestNew2(t *testing.T) {
	New("debug.log", "warn")

	Debugf("hello world")
	Debugf("you name:%s", "duyayun")
	Warnf("warn")
	Infof("info")
}
