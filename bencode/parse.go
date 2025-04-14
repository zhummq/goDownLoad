package bencode

import (
	"bufio"
	"bytes"
	"io"
	"strconv"
)

func parse(data *bufio.Reader) (interface{}, error) {
	c, err := data.ReadByte()
	if err != nil {
		return nil, err
	}
	switch c {
	case 'i':
		integerString, err := optimisticReadBytes(data, 'e')
		if err != nil {
			return nil, err
		}
		integerString = integerString[:len(integerString)-1]
		integer, err := strconv.ParseInt(string(integerString), 10, 64)
		return integer, err
	case 'l':
		list := make([]interface{}, 0)
		for {
			c, err := data.ReadByte()
			if err != nil {
				return nil, err
			}
			if c == 'e' {
				return list, nil
			}
			data.UnreadByte()
			element, err := parse(data)
			if err != nil {
				return nil, err
			}
			list = append(list, element)
		}
	case 'd':
		dict := make(map[string]interface{})
		for {
			c, err := data.ReadByte()
			if err == nil {
				if c == 'e' {
					return dict, nil
				} else {
					data.UnreadByte()
				}
			}
			value, err := parse(data)
			if err != nil {
				return nil, err
			}
			key, ok := value.(string)
			if !ok {
				return nil, err
			}

			value, err = parse(data)
			if err != nil {
				return nil, err
			}
			dict[key] = value
		}
	default:
		data.UnreadByte()
		lengthString, err := optimisticReadBytes(data, ':')
		if err != nil {
			return nil, err
		}
		lengthString = lengthString[:len(lengthString)-1]
		length, err := strconv.ParseInt(string(lengthString), 10, 64)
		if err != nil {
			return nil, err
		}
		buffer := make([]byte, length)
		_, err = io.ReadFull(data, buffer)
		if err != nil {
			return nil, err
		}
		return string(buffer), nil
	}

}

func optimisticReadBytes(data *bufio.Reader, delim byte) ([]byte, error) {
	buffered := data.Buffered()
	var buffer []byte
	var err error
	if buffer, err = data.Peek(buffered); err != nil {
		return nil, err
	}

	if i := bytes.IndexByte(buffer, delim); i >= 0 {
		return data.ReadSlice(delim)
	}
	return data.ReadBytes(delim)
}
