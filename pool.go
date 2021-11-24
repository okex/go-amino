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

func GetBuffer() *bytes.Buffer {
	return bufferPool.Get().(*bytes.Buffer)
}

func PutBuffer(b *bytes.Buffer) {
	b.Reset()
	bufferPool.Put(b)
}
