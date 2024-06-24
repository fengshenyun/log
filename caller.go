package logrec

import (
	"runtime"
	"sync"
)

var (
	m sync.Map
)

func Caller(skip int) (pc uintptr, file string, line int, ok bool) {
	rpc := [1]uintptr{}
	n := runtime.Callers(skip+1, rpc[:])
	if n < 1 {
		return
	}

	pc = rpc[0]
	var frame runtime.Frame
	if item, exist := m.Load(pc); exist {
		frame = item.(runtime.Frame)
	} else {
		tmpPc := []uintptr{pc}
		frame, _ = runtime.CallersFrames(tmpPc).Next()
		m.Store(pc, frame)
	}

	return frame.PC, frame.File, frame.Line, frame.PC != 0
}
