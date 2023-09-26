package serialization

import "errors"

var (
	ErrArgsNil     = errors.New("args is nil")
	ErrNotProtoMsg = errors.New("m must be proto.Message")
)

type Serializer interface {
	Unmarshal(b []byte, m any) error
	Marshal(m any) ([]byte, error)
}
