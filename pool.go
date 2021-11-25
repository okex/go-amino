package amino

import (
	"bytes"
	"sync"
)

var bufferPool = &sync.Pool{
	New: func() interface{} {
		return &bytes.Buffer{}
	},
}

// GetBuffer returns a new bytes.Buffer from the pool.
// you must call PutBuffer on the buffer when you are done with it.
func GetBuffer() *bytes.Buffer {
	return bufferPool.Get().(*bytes.Buffer)
}

// PutBuffer returns a bytes.Buffer to the pool.
func PutBuffer(b *bytes.Buffer) {
	b.Reset()
	bufferPool.Put(b)
}
