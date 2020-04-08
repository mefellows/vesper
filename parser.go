package vesper

import (
	"context"
	"errors"
	"fmt"

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
