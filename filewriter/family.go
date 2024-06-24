package filewriter

import (
	"sync"
)

var (
	defaultWriters = NewFileWriterFamily()
)

type Family struct {
	family map[string]*FileWriter
	mu     sync.RWMutex
}

func NewFileWriterFamily() *Family {
	return &Family{
		family: make(map[string]*FileWriter),
	}
}

func (f *Family) GetOrCreate(fileName string, maxSize int64, maxNum int, maxDay int) (*FileWriter, error) {
	f.mu.RLock()
	w, ok := f.family[fileName]
	f.mu.RUnlock()
	if ok {
		return w, nil
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	w, ok = f.family[fileName]
	if ok {
		return w, nil
	}

	w, err := NewFileWriter(fileName, maxSize, maxNum, maxDay)
	if err == nil {
		f.family[fileName] = w
	}

	return w, err
}
