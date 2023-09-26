package serialization

import (
	"encoding/json"
)

type Json struct{}

var _ Serializer = &Json{}

func (j *Json) Unmarshal(b []byte, m any) error {
	if b == nil || m == nil {
		return ErrArgsNil
	}
	return json.Unmarshal(b, m)
}

func (j *Json) Marshal(m any) ([]byte, error) {
	return json.Marshal(m)
}
