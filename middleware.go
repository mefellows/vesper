package vesper

import (
	"context"
	"encoding/json"
	"log"
	"reflect"
)

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

// newMiddlewareWrapper takes the middleware chain, and converts it into
// a Lambda-compatible interface
func newMiddlewareWrapper(handlerInterface interface{}, middlewareChain LambdaFunc) lambdaHandler {
	handlerType, err := handlerType(handlerInterface)
	if err != nil {
		return errorHandler(err)
	}
	takesContext := handlerTakesContext(handlerType)

	return func(ctx context.Context, payload []byte) (interface{}, error) {
		log.Println("[newMiddlewareWrapper] wrapped function handler")

		ctx = context.WithValue(ctx, ctxKeyPayload, payload)
		var event interface{}
		if (!takesContext && handlerType.NumIn() == 1) || handlerType.NumIn() == 2 {
			eventType := handlerType.In(handlerType.NumIn() - 1)
			ctx = context.WithValue(ctx, ctxKeyTIn, eventType)
			evt := reflect.New(eventType)

			if err := json.Unmarshal(payload, evt.Interface()); err != nil {
				return nil, err
			}
			event = evt.Elem().Interface()
		}

		return middlewareChain(ctx, event)
	}
}

// newTypedToUntypedWrapper takes a typed handler function
// and converts it into a Middleware-compatible function
func newTypedToUntypedWrapper(handlerInterface interface{}) LambdaFunc {
	errLambdaFunc := func(err error) func(context.Context, interface{}) (interface{}, error) {
		return func(context.Context, interface{}) (interface{}, error) {
			return nil, err
		}
	}

	if err := validateHandlerFunc(handlerInterface); err != nil {
		return errLambdaFunc(err)
	}
	handlerType, err := handlerType(handlerInterface)
	if err != nil {
		return errLambdaFunc(err)
	}
	takesContext := handlerTakesContext(handlerType)
	handler := reflect.ValueOf(handlerInterface)

	return func(ctx context.Context, payload interface{}) (interface{}, error) {
		log.Printf("[typedToUntypedWrapper] have payload: %+v \n", payload)

		// construct arguments
		var args []reflect.Value
		if takesContext {
			args = append(args, reflect.ValueOf(ctx))
		}
		if (!takesContext && handlerType.NumIn() == 1) || handlerType.NumIn() == 2 {
			args = append(args, reflect.ValueOf(payload))
		}

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
