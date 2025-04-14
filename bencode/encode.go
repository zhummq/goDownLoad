package bencode

import (
	"bytes"
	"reflect"
	"strconv"
)

func Encode(v interface{}) ([]byte, error) {
	var buf bytes.Buffer
	err := encodeValue(&buf, reflect.ValueOf(v))
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func encodeValue(buf *bytes.Buffer, v reflect.Value) error {
	switch v.Kind() {
	case reflect.String:
		encodeString(buf, v.String())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		encodeInt(buf, v.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		encodeUint(buf, v.Uint())
	case reflect.Slice, reflect.Array:
		return encodeList(buf, v)
	case reflect.Map:
		return encodeDict(buf, v)
	case reflect.Interface, reflect.Ptr:
		return encodeValue(buf, v.Elem())
	case reflect.Bool:
		if v.Bool() {
			encodeInt(buf, 1)
		} else {
			encodeInt(buf, 0)
		}
	default:
		return nil
	}
	return nil
}

func encodeString(buf *bytes.Buffer, s string) {
	buf.WriteString(strconv.Itoa(len(s)))
	buf.WriteByte(':')
	buf.WriteString(s)
}

// encodeInt 编码整数
func encodeInt(buf *bytes.Buffer, i int64) {
	buf.WriteByte('i')
	buf.WriteString(strconv.FormatInt(i, 10))
	buf.WriteByte('e')
}

// encodeUint 编码无符号整数
func encodeUint(buf *bytes.Buffer, u uint64) {
	buf.WriteByte('i')
	buf.WriteString(strconv.FormatUint(u, 10))
	buf.WriteByte('e')
}

// encodeList 编码列表
func encodeList(buf *bytes.Buffer, v reflect.Value) error {
	buf.WriteByte('l')
	for i := 0; i < v.Len(); i++ {
		if err := encodeValue(buf, v.Index(i)); err != nil {
			return err
		}
	}
	buf.WriteByte('e')
	return nil
}

// encodeDict 编码字典
func encodeDict(buf *bytes.Buffer, v reflect.Value) error {
	buf.WriteByte('d')

	// 获取字典的键值对
	keys := v.MapKeys()
	for _, key := range keys {
		// 编码键（必须是字符串）
		if key.Kind() != reflect.String {
			return nil
		}
		encodeString(buf, key.String())

		// 编码值
		if err := encodeValue(buf, v.MapIndex(key)); err != nil {
			return err
		}
	}

	buf.WriteByte('e')
	return nil
}
