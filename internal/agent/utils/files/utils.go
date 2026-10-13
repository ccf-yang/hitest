package files

import (
	"bufio"
	"sync"
)

const (
	MaxReadFileSize = 512 * 1024 * 1024
	tailBufSize     = int64(32768)
)

const dotCharacter = 46

var readerPool = sync.Pool{
	New: func() interface{} {
		return bufio.NewReaderSize(nil, 8192)
	},
}

var tailBufPool = sync.Pool{
	New: func() interface{} {
		buf := make([]byte, tailBufSize)
		return &buf
	},
}
