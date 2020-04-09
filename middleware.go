package vesper

import (
	"context"
	"fmt"
	"reflect"
)

// LambdaFunc is the long-form of the Lambda handler interface
// https://github.com/aws/aws-lambda-go/blob/master/lambda/entry.go#L37-L49
type LambdaFunc func(context.Context, interface{}) (interface{}, error)

// Middleware is a definition of what a Middleware is,
// take in one LambdaFunc and wrap it within another LambdaFunc
type Middleware func(next LambdaFunc) LambdaFunc

// buildChain builds the middlware chain recursively, functions are first class
func buildChain(f LambdaFunc, m ...Middleware) LambdaFunc {
	// if our chain is done, use the original LambdaFunc
	if len(m) == 0 {
		return f
	}
	// otherwise nest the LambdaFuncs
	return m[0](buildChain(f, m[1:]...))
}

// newMiddlewareWrapper takes the middleware chain, and converts it into
// a Lambda-compatible interface
func newMiddlewareWrapper(handlerInterface interface{}, middlewareChain LambdaFunc) lambdaHandler {
	handlerType, err := handlerType(handlerInterface)
	if err != nil {
		return errorHandler(err)
	}
	takesContext := handlerTakesContext(handlerType)
	var tIn reflect.Type
	if (!takesContext && handlerType.NumIn() == 1) || handlerType.NumIn() == 2 {
		tIn = handlerType.In(handlerType.NumIn() - 1)
	}
	return func(ctx context.Context, payload []byte) (interface{}, error) {
		log.Println("[newMiddlewareWrapper] wrapped function handler")

		ctx = context.WithValue(ctx, ctxKeyPayload, payload)
		ctx = context.WithValue(ctx, ctxKeyTIn, tIn)
		return middlewareChain(ctx, payload)
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

	var tIn reflect.Type
	if (!takesContext && handlerType.NumIn() == 1) || handlerType.NumIn() == 2 {
		tIn = handlerType.In(handlerType.NumIn() - 1)
	}

	return func(ctx context.Context, payload interface{}) (interface{}, error) {
		log.Printf("[typedToUntypedWrapper] have payload: %+v \n", payload)

		// construct arguments
		var args []reflect.Value
		if takesContext {
			args = append(args, reflect.ValueOf(ctx))
		}
		if tIn != nil {
			t := reflect.TypeOf(payload)
			if t != nil && !t.AssignableTo(tIn) {
				return nil, fmt.Errorf("expected payload type of %s but got %s when calling the handler. parser middlewares probably need to be added", tIn.String(), t)
			}
			args = append(args, reflectValueOrZero(tIn, payload))
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

func reflectValueOrZero(tIn reflect.Type, val interface{}) reflect.Value {
	if val == nil {
		return reflect.New(tIn).Elem()
	}
	return reflect.ValueOf(val)
}
