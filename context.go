package vesper

import (
	"context"
	"reflect"
)

type ctxKey string

const (
	ctxKeyPayload = ctxKey("payload")
	ctxKeyTIn     = ctxKey("TIn")
)

// PayloadFromContext retrieves the original payload with type []byte from a context.
func PayloadFromContext(ctx context.Context) ([]byte, bool) {
	value, ok := ctx.Value(ctxKeyPayload).([]byte)
	return value, ok
}

// TInFromContext retrieves the TIn of the lambda handler with type reflect.Type from a context.
func TInFromContext(ctx context.Context) (reflect.Type, bool) {
	value, ok := ctx.Value(ctxKeyTIn).(reflect.Type)
	return value, ok
}
