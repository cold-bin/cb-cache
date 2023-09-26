package serialization

import (
	"github.com/golang/protobuf/proto"
)

type Protobuf struct{}

var _ Serializer = &Protobuf{}

func (p *Protobuf) Unmarshal(b []byte, m any) error {
	if b == nil || m == nil {
		return ErrArgsNil
	}
	if v, ok := m.(proto.Message); ok {
		return proto.Unmarshal(b, v)
	}
	return ErrNotProtoMsg
}

func (p *Protobuf) Marshal(m any) ([]byte, error) {
	if v, ok := m.(proto.Message); ok {
		return proto.Marshal(v)
	}
	return nil, ErrNotProtoMsg
}
