package main

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	// "github.com/aws/aws-lambda-go/lambdacontext"
)

// Typed interface for testing
type foo struct {
	Foo string
}

// MyHandler implements the Lambda Handler interface
func testHandler(ctx context.Context, o foo) (interface{}, error) {
	fmt.Println("[actual handler]: Start: ", ctx, o.Foo)

	return o.Foo, nil
}

// testHandlerMiddleware is a test middleware
type testHandlerMiddleware struct{}

// Handle implements the handler interface.Handle
// TODO: allow this to be a function so we can simplify the interface
// TODO: allow this function to accept _any_ type as `in` and reflect on it to ensure the correct shape.
//       this will be nice for
func (m testHandlerMiddleware) Handle(ctx context.Context, in interface{}, next LambdaFunc) (interface{}, error) {
	fmt.Println("[testHandlerMiddleware] START: ", ctx, in)
	newIn := in.(foo)
	if newIn.Foo == "bar" {
		newIn.Foo = "not bar!"
	}
	if newIn.Foo == "cached" {
		fmt.Println("Cached response found, bailing out of middleware!")
		newIn.Foo = "not bar!"
		return "cached result man!", nil
	}
	res, err := next(ctx, in)

	fmt.Println("[testHandlerMiddleware] END: ", ctx, in)
	return res, err
}

// mw is a definition of what a mw is,
// take in one LambdaFunc and wrap it within another LambdaFunc
type mw func(LambdaFunc) LambdaFunc

// func setup(h interface{}, m ...mw) LambdaFunc {
// 	handler := newHandler(h)
// 	return buildChain(handler, m...)
// }

// buildChain builds the middlware chain recursively, functions are first class
func buildChain(f LambdaFunc, m ...mw) LambdaFunc {
	// if our chain is done, use the original LambdaFunc
	if len(m) == 0 {
		return f
	}
	// otherwise nest the LambdaFuncs
	return m[0](buildChain(f, m[1:cap(m)]...))
}

var TypeMiddleware = func(i interface{}) mw {
	// TODO: scream if TypeOf returns a primitive?
	out := reflect.New(reflect.TypeOf(i))

	return func(f LambdaFunc) LambdaFunc {
		return func(ctx context.Context, in interface{}) (interface{}, error) {
			fmt.Println("[typeMiddleware] start: ", ctx, in)
			switch t := in.(type) {
			case []byte:
				fmt.Println("[typeMiddleware] byte array, converting to object of type ", reflect.TypeOf(i))
				if err := json.Unmarshal(t, out.Interface()); err != nil {
					return nil, err
				}
				return f(ctx, out.Elem().Interface())
			case string:
				fmt.Println("[typeMiddleware] string, converting to object of type ", t, reflect.TypeOf(i))
				if err := json.Unmarshal([]byte(t), out.Interface()); err != nil {
					return nil, err
				}
				return f(ctx, out.Elem().Interface())
			}

			fmt.Println("[typeMiddleware] END: ")
			return f(ctx, in)
		}
	}
}

var AuthMiddleware = func(f LambdaFunc) LambdaFunc {
	// one time scope setup area for middleware

	return func(ctx context.Context, in interface{}) (interface{}, error) {
		fmt.Println("[authMiddleware] start: ", ctx, in)
		res, err := f(ctx, in)
		fmt.Println("[authMiddleware] END: ", ctx, in)

		return res, err
	}
}
var DummyMiddleware = func(f LambdaFunc) LambdaFunc {
	// one time scope setup area for middleware

	return func(ctx context.Context, in interface{}) (interface{}, error) {
		fmt.Println("[dummyMiddleware] start: ", ctx, in)
		res, err := f(ctx, in)
		fmt.Println("[dummyMiddleware] END: ", ctx, in)

		return res, err
	}
}

// chainHandler implements the Lambda Handler interface
var chainHandler = func(ctx context.Context, i interface{}) (interface{}, error) {
	fmt.Println("[actual handler]: Start: ", ctx, i)
	newIn := i.(foo)
	fmt.Println(newIn)
	fmt.Printf("[chainHandler] received: %+v \n\n", i)

	return newIn.Foo, nil
}

type lambdaHandler func(context.Context, []byte) (interface{}, error)

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

func validateArguments(handler reflect.Type) (bool, error) {
	handlerTakesContext := false
	if handler.NumIn() > 2 {
		return false, fmt.Errorf("handlers may not take more than two arguments, but handler takes %d", handler.NumIn())
	} else if handler.NumIn() > 0 {
		contextType := reflect.TypeOf((*context.Context)(nil)).Elem()
		argumentType := handler.In(0)
		handlerTakesContext = argumentType.Implements(contextType)
		if handler.NumIn() > 1 && !handlerTakesContext {
			return false, fmt.Errorf("handler takes two arguments, but the first is not Context. got %s", argumentType.Kind())
		}
	}

	return handlerTakesContext, nil
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

// newHandler Creates the base lambda handler, which will do basic payload unmarshaling before defering to handlerSymbol.
// If handlerSymbol is not a valid handler, the returned function will be a handler that just reports the validation error.
func newHandler(handlerSymbol interface{}) lambdaHandler {
	if handlerSymbol == nil {
		return errorHandler(fmt.Errorf("handler is nil"))
	}
	handler := reflect.ValueOf(handlerSymbol)
	handlerType := reflect.TypeOf(handlerSymbol)
	if handlerType.Kind() != reflect.Func {
		return errorHandler(fmt.Errorf("handler kind %s is not %s", handlerType.Kind(), reflect.Func))
	}

	takesContext, err := validateArguments(handlerType)
	if err != nil {
		return errorHandler(err)
	}

	if err := validateReturns(handlerType); err != nil {
		return errorHandler(err)
	}

	return func(ctx context.Context, payload []byte) (interface{}, error) {
		// construct arguments
		var args []reflect.Value
		if takesContext {
			args = append(args, reflect.ValueOf(ctx))
		}
		if (handlerType.NumIn() == 1 && !takesContext) || handlerType.NumIn() == 2 {
			eventType := handlerType.In(handlerType.NumIn() - 1)
			event := reflect.New(eventType)

			if err := json.Unmarshal(payload, event.Interface()); err != nil {
				return nil, err
			}

			args = append(args, event.Elem())
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

func main() {
	// m := buildChain(chainHandler, TypeMiddleware(foo{}), AuthMiddleware)
	// lambda.Start(m)
}
