package encoding

type UnmarshalFunc func(data []byte, v interface{}) error
