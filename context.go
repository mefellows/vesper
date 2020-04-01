package vesper

import (
	"context"
	"reflect"
)

type ctxKey string

const (
	ctxKeyPayload          = ctxKey("payload")
	ctxKeyHandlerSignature = ctxKey("handlerSignature")
)

// PayloadFromContext retrieves the original payload with type []byte from a context.
func PayloadFromContext(ctx context.Context) ([]byte, bool) {
	value, ok := ctx.Value(ctxKeyPayload).([]byte)
	return value, ok
}

// HandlerSignatureFromContext retrieves the Lambda handler with type reflect.Type from a context.
func HandlerSignatureFromContext(ctx context.Context) (reflect.Type, bool) {
	value, ok := ctx.Value(ctxKeyHandlerSignature).(reflect.Type)
	return value, ok
}
