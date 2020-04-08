package vesper

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/mefellows/vesper/encoding"
)

func ParserMiddleware(unmarshaler encoding.UnmarshalFunc) func(LambdaFunc) LambdaFunc {
	return func(next LambdaFunc) LambdaFunc {
		return func(ctx context.Context, in interface{}) (interface{}, error) {
			payload, ok := PayloadFromContext(ctx)
			if !ok {
				return nil, errors.New("payload not available in context. this is likely a bug with the framework")
			}
			tIn, ok := TInFromContext(ctx)
			if !ok {
				return next(ctx, in) // continue as there is no TIn to parse anyway.
			}
			nextIn, err := unmarshalToType(unmarshaler, tIn, payload)
			if err != nil {
				return nil, fmt.Errorf("could not unmarshal payload to type of '%s': %w", tIn.String(), err)
			}
			return next(ctx, nextIn)
		}
	}
}

func JSONParserMiddleware() func(LambdaFunc) LambdaFunc {
	return ParserMiddleware(encoding.JSONUnmarshal)
}

func unmarshalToType(unmarshal encoding.UnmarshalFunc, t reflect.Type, payload []byte) (interface{}, error) {
	evt := reflect.New(t)
	if err := unmarshal(payload, evt.Interface()); err != nil {
		return nil, err
	}
	return evt.Elem().Interface(), nil
}
