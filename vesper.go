package vesper

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"reflect"

	"github.com/aws/aws-lambda-go/lambda"
)

// Vesper is a middleware adapter for Lambda Functions
type Vesper struct {
	handler lambdaHandler
	log     *log.Logger
}

// New creates a new Vesper instance given a Handler and set of Middleware
// Middlewares are evaluated in the order they are provided
func New(l interface{}, middlewares ...Middleware) *Vesper {
	m := buildChain(newTypedToUntypedWrapper(l), middlewares...)
	f := newMiddlewareWrapper(l, m)

	return &Vesper{
		handler: f,
		log:     log.New(os.Stdout, "", log.LstdFlags),
	}
}

// Start is a convenience function run the lambda handler
func (v *Vesper) Start() {
	lambda.StartHandler(v.handler)
}

// WithLogger sets the log instance to use
func (v *Vesper) WithLogger(log *log.Logger) {
	v.log = log
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
