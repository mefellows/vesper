package vesper

import (
	"context"
	"reflect"
)

type ctxKey string

const (
	ctxKeyPayload = ctxKey("payload")
	ctxKeyTIn     = ctxKey("TIn")
	ctxKeyEvent   = ctxKey("event")
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

// EventFromContext retrieves the event which triggered the lambda from a context.
// This is only set if you used one of the event parser middlewares (e.g. SQSParser).
// Events are with the types found in this package github.com/aws/aws-lambda-go/events.
func EventFromContext(ctx context.Context) (reflect.Type, bool) {
	value, ok := ctx.Value(ctxKeyTIn).(reflect.Type)
	return value, ok
}
