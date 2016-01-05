package dbutil

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"strconv"
)

func MustInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}

	return i
}

func ConvertString(inter interface{}) string {
	switch v := inter.(type) {
	case string:
		return v
	case float64:
		return strconv.FormatFloat(v, 'f', 0, 64)
	case int64:
		return strconv.FormatInt(v, 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	case int:
		return strconv.Itoa(v)
	default:
		// 认为出错了，返回""
		return ""
	}
}

func Md5(text string) string {
	hashMd5 := md5.New()
	io.WriteString(hashMd5, text)
	return fmt.Sprintf("%x", hashMd5.Sum(nil))
}

// 内嵌bytes.Buffer，支持连写
type Buffer struct {
	*bytes.Buffer
}

func NewBuffer() *Buffer {
	return &Buffer{Buffer: new(bytes.Buffer)}
}

func (b *Buffer) Append(s string) *Buffer {
	defer func() {
		if err := recover(); err != nil {
			log.Println("*****内存不够了！******")
		}
	}()

	b.WriteString(s)
	return b
}

func (b *Buffer) AppendInt(i int64) *Buffer {
	return b.Append(strconv.FormatInt(i, 10))
}

func (b *Buffer) AppendUint(i uint64) *Buffer {
	return b.Append(strconv.FormatUint(i, 10))
}

func (b *Buffer) AppendBytes(p []byte) *Buffer {
	defer func() {
		if err := recover(); err != nil {
			log.Println("*****内存不够了！******")
		}
	}()

	b.Write(p)

	return b
}

func (b *Buffer) AppendRune(r rune) *Buffer {
	defer func() {
		if err := recover(); err != nil {
			log.Println("*****内存不够了！******")
		}
	}()

	b.WriteRune(r)

	return b
}
