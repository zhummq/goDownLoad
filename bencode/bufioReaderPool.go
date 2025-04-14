package bencode

import (
	"bufio"
	"io"
	"sync"
)

var bufioReaderPool sync.Pool

func getBufioReader(Reader io.Reader) *bufio.Reader {
	v := bufioReaderPool.Get()
	if v == nil {
		return bufio.NewReader(Reader)
	}
	r := v.(*bufio.Reader)
	r.Reset(Reader)
	return r
}
