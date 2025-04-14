package bencode

import (
	"bufio"
	"io"
)

func Decode(reader io.Reader) (interface{}, error) {
	bufioReader, ok := reader.(*bufio.Reader)
	if !ok {
		bufioReader = getBufioReader(reader)
		defer bufioReaderPool.Put(bufioReader)
	}
	return parse(bufioReader)

}
