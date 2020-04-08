package vesper

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/aws/aws-lambda-go/lambda"
)

var log logPrinter = noOpPrinter{}

// Vesper is a middleware adapter for Lambda Functions
type Vesper struct {
	rawHandler    interface{}
	middlewares   []Middleware
	autoUnmarshal bool
}

// New creates a new Vesper instance given a Handler and set of Middleware
// Middlewares are evaluated in the order they are provided
func New(handler interface{}, middlewares ...Middleware) *Vesper {
	v := Vesper{
		rawHandler:    handler,
		middlewares:   middlewares,
		autoUnmarshal: true,
	}
	return &v
}

func (v *Vesper) DisableAutoUnmarshal() *Vesper {
	v.autoUnmarshal = false
	return v
}

func (v *Vesper) AddMiddlewares(middlewares ...Middleware) *Vesper {
	v.middlewares = append(v.middlewares, middlewares...)
	return v
}

func (v *Vesper) AddMiddleware(middleware Middleware) *Vesper {
	return v.AddMiddlewares(middleware)
}

func (v *Vesper) buildHandler() lambdaHandler {
	mids := v.middlewares
	if v.autoUnmarshal {
		mids = append([]Middleware{JSONParserMiddleware()}, mids...)
	}
	m := buildChain(newTypedToUntypedWrapper(v.rawHandler), mids...)
	return newMiddlewareWrapper(v.rawHandler, m)
}

// Start is a convenience function run the lambda handler
func (v *Vesper) Start() {
	lambda.StartHandler(v.buildHandler())
}

// Logger sets the log to use
func Logger(l logPrinter) {
	log = l
}

// ExtractType fetches the original invocation payload (as a []byte)
// and converts it to the given narrow type
// This is useful for situations where a function is invoked from multiple
// contexts (e.g. warmup, http, S3 events) and handlers/middlewares need to be strongly
// typed
func ExtractType(ctx context.Context, in interface{}) error {
	t := reflect.TypeOf(in)

	if t != nil && t.Name() != "interface" {
		if v, ok := PayloadFromContext(ctx); ok {
			err := json.Unmarshal(v, &in)
			if err != nil {
				return extractError(t.Name(), err)
			}
			return nil
		}
		return extractError(t.Name(), nil)
	}

	return extractError(t.Name(), nil)
}

func extractError(t string, e error) error {
	return fmt.Errorf("unable to narrow type to %v: %v", t, e)
}
