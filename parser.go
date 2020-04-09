package vesper

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/mefellows/vesper/encoding"
)

// ParserMiddleware is a middleware which unmarshals the original payload to the handler input parameter type with the given unmarshaler.
func ParserMiddleware(unmarshaler encoding.UnmarshalFunc) func(LambdaFunc) LambdaFunc {
	return func(next LambdaFunc) LambdaFunc {
		return func(ctx context.Context, in interface{}) (interface{}, error) {
			if unmarshaler == nil {
				return nil, errors.New("no unmarshaler was provided")
			}
			b, ok := in.([]byte)
			if !ok {
				return nil, fmt.Errorf("parser middleware expected []byte input but got %T", in)
			}
			tIn, ok := TInFromContext(ctx)
			if !ok {
				return next(ctx, in) // continue as there is no TIn to parse anyway.
			}
			nextIn, err := unmarshalToType(unmarshaler, tIn, b)
			if err != nil {
				return nil, fmt.Errorf("could not unmarshal payload to type of '%s': %w", tIn.String(), err)
			}
			return next(ctx, nextIn)
		}
	}
}

// JSONParserMiddleware is a middleware which JSON unmarshals the original payload to the handler input parameter type.
func JSONParserMiddleware() func(LambdaFunc) LambdaFunc {
	return ParserMiddleware(json.Unmarshal)
}

func unmarshalToType(unmarshal encoding.UnmarshalFunc, t reflect.Type, payload []byte) (interface{}, error) {
	evt := reflect.New(t)
	if err := unmarshal(payload, evt.Interface()); err != nil {
		return nil, err
	}
	return evt.Elem().Interface(), nil
}
