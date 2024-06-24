package filewriter

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWrite01(t *testing.T) {
	w, err := NewFileWriter("log/test01", 1024, 10, 1)
	assert.Equal(t, nil, err)

	for i := 0; i < 100; i++ {
		w.Write([]byte("abcdefghijklmnopqrstuvwxyz"))
	}

	w.Flush()
	fmt.Println("writer stopped!")
}

/*
func TestRotate(t *testing.T) {
	w, err := NewFileWriter("log/test02", 1024, 10, 1)
	assert.Equal(t, nil, err)

	for i := 0; i < 100; i++ {
		w.Write([]byte("abcdefghijklmnopqrstuvwxyz"))
	}
	w.Flush()

	time.Sleep(2 * time.Minute)

	w.Write([]byte("0123456789"))
	w.Flush()
}
*/

func TestGetFileWriter(t *testing.T) {
	fw, err := GetFileWriter("log/test03", 2048, 10, 1)
	assert.Equal(t, nil, err)

	for i := 0; i < 100; i++ {
		fw.Write([]byte("abcdefghijklmnopqrstuvwxyz"))
	}
	fw.Write([]byte("\n"))
	fw.Flush()

	fw1, err1 := GetFileWriter("log/test03", 2048, 10, 1)
	assert.Equal(t, nil, err1)
	assert.Equal(t, fw, fw1)

	fw1.Write([]byte("0123456789"))
	fw.Flush()
}
