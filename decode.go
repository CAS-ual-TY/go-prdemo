package prdemo

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"io"
	"os"
)

var (
	demoEndian = binary.LittleEndian
)

type ReadAtSeeker interface {
	io.ReaderAt
	io.ReadSeeker
}

type Decoder interface {
	Decode(*Message) error
}

type DemoReader interface {
	Next() bool
	GetMessage() (*Message, error)
}

type demoReader struct {
	r          ReadAtSeeker
	pos        int64
	size       int64
	nextMsgLen uint16
}

func NewDemoReader(r ReadAtSeeker) (DemoReader, error) {
	current, err := r.Seek(0, io.SeekCurrent)
	if err != nil {
		return nil, err
	}

	end, err := r.Seek(0, io.SeekEnd)
	if err != nil {
		return nil, err
	}

	_, err = r.Seek(current, io.SeekStart)
	if err != nil {
		return nil, err
	}

	return &demoReader{
		r:    r,
		size: end,
	}, nil
}

func NewDemoReaderFromFile(file string) (DemoReader, error) {
	reader, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	zReader, err := zlib.NewReader(reader)
	if err != nil {
		return nil, err
	}

	buf, err := io.ReadAll(zReader)
	if err != nil {
		return nil, err
	}

	return NewDemoReader(bytes.NewReader(buf))
}

func (r *demoReader) Next() bool {
	var len uint16

	err := binary.Read(r.r, demoEndian, &len)
	if err != nil {
		return false
	}

	r.pos += 2

	if r.pos+int64(len) > r.size {
		return false
	}

	r.nextMsgLen = len

	return true
}

func (r *demoReader) GetMessage() (*Message, error) {
	return NewMessage(io.NewSectionReader(r.r, r.pos, int64(r.nextMsgLen)))
}
