package encoding

import "encoding/json"

type UnmarshalFunc func(data []byte, v interface{}) error

func JSONUnmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
