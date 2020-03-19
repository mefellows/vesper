package vesper

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
)

// PAYLOAD is the context key to retrieve the original request payload
// as a []byte
type PAYLOAD struct{}

// LambdaFunc is the long-form of the Lambda handler interface
// https://github.com/aws/aws-lambda-go/blob/master/lambda/entry.go#L37-L49
type LambdaFunc func(context.Context, interface{}) (interface{}, error)

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

func typedMiddleware(m interface{}) LambdaFunc {
	// Extract out first parameter

	// wrap in untyped interface (call existing wrapper?)

	handler := reflect.ValueOf(m)
	log.Println("[**typedMiddleware] ", handler)

	return func(ctx context.Context, payload interface{}) (interface{}, error) {
		log.Printf("[**typedMiddleware] have payload: %+v \n", payload)

		// construct arguments
		var args []reflect.Value
		args = append(args, reflect.ValueOf(ctx))
		args = append(args, reflect.ValueOf(payload))

		return nil, nil
	}
}

// newMiddlewareWrapper takes the middleware chain, and converts it into
// a Lambda-compatible interface
func newMiddlewareWrapper(handlerInterface interface{}, middlewareChain LambdaFunc) lambdaHandler {
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
		log.Println("[newMiddlewareWrapper] wrapped function handler")
		newCtx := context.WithValue(ctx, PAYLOAD{}, payload)
		eventType := handlerType.In(handlerType.NumIn() - 1)
		event := reflect.New(eventType)

		if err := json.Unmarshal(payload, event.Interface()); err != nil {
			return nil, err
		}

		return middlewareChain(newCtx, event.Elem().Interface())
	}
}

// newTypedToUntypedWrapper takes a typed handler function
// and converts it into a Middleware-compatible function
func newTypedToUntypedWrapper(handlerInterface interface{}) LambdaFunc {
	handler := reflect.ValueOf(handlerInterface)

	return func(ctx context.Context, payload interface{}) (interface{}, error) {
		log.Printf("[typedToUntypedWrapper] have payload: %+v \n", payload)

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
