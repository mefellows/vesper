package vesper

import (
	"os"

	"log"

	"github.com/aws/aws-lambda-go/lambda"
)

// Vesper is a middleware adapter for Lambda Functions
// Middleware are evaluated in the order they are added to the stack
type Vesper struct {
	handler lambdaHandler
	log     *log.Logger
}

// New creates a new Vesper instance given a Handler and set of Middleware
func New(l interface{}, handlers ...Middleware) *Vesper {
	m := buildChain(newTypedToUntypedWrapper(l), handlers...)
	f := newMiddlewareWrapper(l, m)

	return &Vesper{
		handler: f,
		log:     log.New(os.Stdout, "", log.LstdFlags),
	}
}

// Start is a convienence function run the lambda handler
func (v *Vesper) Start() {
	lambda.StartHandler(v.handler)
}

// WithLogger sets the log instance to use
func (v *Vesper) WithLogger(log *log.Logger) {
	v.log = log
}
