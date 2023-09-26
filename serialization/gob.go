package serialization

import (
	"bytes"
	"encoding/gob"
	"sync"
)

type Gob struct {
	buf *bytes.Buffer
}

var _ Serializer = &Gob{}

var bufPool sync.Pool

func init() {
	bufPool.New = func() any {
		return bytes.NewBuffer([]byte{})
	}
}

func getBuf() *bytes.Buffer {
	return bufPool.Get().(*bytes.Buffer)
}

func putBuf(buf *bytes.Buffer) {
	if buf == nil {
		buf = bytes.NewBuffer([]byte{})
	}
	buf.Reset()
	bufPool.Put(buf)
}

func (g *Gob) Unmarshal(b []byte, m any) error {
	if b == nil || m == nil {
		return ErrArgsNil
	}
	g.buf = getBuf()
	g.buf.Write(b)
	return gob.NewDecoder(g.buf).Decode(m)
}

func (g *Gob) Marshal(m any) ([]byte, error) {
	g.buf = getBuf()
	defer putBuf(g.buf)

	if err := gob.NewEncoder(g.buf).Encode(m); err != nil {
		return nil, err
	}

	// deep copy
	res := make([]byte, g.buf.Len())
	copy(res, g.buf.Bytes())

	return res, nil
}
