package vesper

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/aws/aws-lambda-go/lambda"
)

// TODO: 1. Support passing in functions, instead of structs - DONE
//       2. Support modification of request/response (e.g. pass through objects?)
//       3. ...WithContext handler wrapper? -> dah, already in the interface!! - DONE
//       4. Logging
//       5. Confirm what types Lambda accepts in its interface - DONE (not primitives!)

// LambdaFunc is (long-form) of the Lambda handler interface
// https://github.com/aws/aws-lambda-go/blob/master/lambda/entry.go#L37-L49
type LambdaFunc func(context.Context, interface{}) (interface{}, error)

type lambdaHandler func(context.Context, []byte) (interface{}, error)

// Middleware is a definition of what a Middleware is,
// take in one LambdaFunc and wrap it within another LambdaFunc
type Middleware func(LambdaFunc) LambdaFunc

// buildChain builds the middlware chain recursively, functions are first class
func buildChain(f LambdaFunc, m ...Middleware) LambdaFunc {
	// if our chain is done, use the original LambdaFunc
	if len(m) == 0 {
		return f
	}
	// otherwise nest the LambdaFuncs
	return m[0](buildChain(f, m[1:cap(m)]...))
}

// Invoke calls the handler, and serializes the response.
// If the underlying handler returned an error, or an error occurs during serialization, error is returned.
func (handler lambdaHandler) Invoke(ctx context.Context, payload []byte) ([]byte, error) {
	response, err := handler(ctx, payload)
	if err != nil {
		return nil, err
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		return nil, err
	}

	return responseBytes, nil
}

func errorHandler(e error) lambdaHandler {
	return func(ctx context.Context, event []byte) (interface{}, error) {
		return nil, e
	}
}

func validateReturns(handler reflect.Type) error {
	errorType := reflect.TypeOf((*error)(nil)).Elem()
	if handler.NumOut() > 2 {
		return fmt.Errorf("handler may not return more than two values")
	} else if handler.NumOut() > 1 {
		if !handler.Out(1).Implements(errorType) {
			return fmt.Errorf("handler returns two values, but the second does not implement error")
		}
	} else if handler.NumOut() == 1 {
		if !handler.Out(0).Implements(errorType) {
			return fmt.Errorf("handler returns a single value, but it does not implement error")
		}
	}
	return nil
}

func newMiddlewareWrapper(handlerInterface interface{}, middlewareChain LambdaFunc) lambdaHandler {
	fmt.Println("[newMiddlewareWrapper] wrapping middleware in type-preserving, Lambda-compatible shell")

	if handlerInterface == nil {
		return errorHandler(fmt.Errorf("handler is nil"))
	}
	handlerType := reflect.TypeOf(handlerInterface)
	if handlerType.Kind() != reflect.Func {
		return errorHandler(fmt.Errorf("handler kind %s is not %s", handlerType.Kind(), reflect.Func))
	}

	if err := validateReturns(handlerType); err != nil {
		return errorHandler(err)
	}

	return func(ctx context.Context, payload []byte) (interface{}, error) {
		eventType := handlerType.In(handlerType.NumIn() - 1)
		event := reflect.New(eventType)

		if err := json.Unmarshal(payload, event.Interface()); err != nil {
			return nil, err
		}

		// TODO: Rather than call the original handler, we invoke the middleware chain
		//       This way we have the signature that AWS needs
		// response := handler.Call(args)
		return middlewareChain(ctx, event.Elem().Interface())
	}
}

func newTypedToUntypedWrapper(handlerInterface interface{}) LambdaFunc {
	fmt.Println("[typedToUntypedWrapper] wrapping typed handler interface in MW compatible shell")
	handler := reflect.ValueOf(handlerInterface)

	return func(ctx context.Context, payload interface{}) (interface{}, error) {
		fmt.Printf("[typedToUntypedWrapper] have payload: %+v \n", payload)

		// construct arguments
		var args []reflect.Value
		args = append(args, reflect.ValueOf(ctx))
		args = append(args, reflect.ValueOf(payload))

		response := handler.Call(args)

		// convert return values into (interface{}, error)
		var err error
		if len(response) > 0 {
			if errVal, ok := response[len(response)-1].Interface().(error); ok {
				err = errVal
			}
		}
		var val interface{}
		if len(response) > 1 {
			val = response[0].Interface()
		}

		return val, err
	}
}

// Middy is a middleware adapter for Lambda Functions
// Middleware are evaluated in the order they are added to the stack
type Middy struct {
	handler lambdaHandler
}

// New creates a new Middy given a set of Handlers
func New(l interface{}, handlers ...Middleware) *Middy {
	m := buildChain(newTypedToUntypedWrapper(l), handlers...)
	f := newMiddlewareWrapper(l, m)

	return &Middy{
		handler: f,
	}
}

// Start is a convienence function run the lambda handler
func (m *Middy) Start() {
	lambda.StartHandler(m.handler)
}
