package main

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

// yetAnotherMiddleware is a test middleware
type yetAnotherMiddleware struct{}

func (m yetAnotherMiddleware) Handle(ctx context.Context, in interface{}, next LambdaFunc) (interface{}, error) {
	fmt.Println("[yetAnotherMiddleware] start: ", ctx, in)
	res, err := next(ctx, in)
	fmt.Println("[yetAnotherMiddleware] END: ", ctx, in)

	return res, err
}

func Test_middleware(t *testing.T) {
	fmt.Println("Testing")
	ctx := context.TODO()
	i := `{"foo":"bar"}`
	m := New(testHandler, testHandlerMiddleware{}, yetAnotherMiddleware{})
	res, err := m.Handle()(ctx, i)
	fmt.Printf("res: %v, err: %v", res, err)
}

func Test_buildMiddleware(t *testing.T) {
	m := buildChain(chainHandler, TypeMiddleware(foo{}), AuthMiddleware)
	ctx := context.TODO()

	// The Lamda handler interface takes a context and a byte[]
	i := []byte(`{"foo":"bar"}`)
	res, err := m(ctx, i)
	fmt.Printf("res: %v, err: %v", res, err)
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

func Test_buildTypedMiddleware(t *testing.T) {
	m := buildChain(newTypedToUntypedWrapper(testHandler), AuthMiddleware, DummyMiddleware)
	f := newMiddlewareWrapper(testHandler, m)
	ctx := context.TODO()

	// The Lamda handler interface takes a context and a byte[]
	i := []byte(`{"foo":"bar"}`)
	res, err := f(ctx, i)

	fmt.Printf("res: %v, err: %v", res, err)
}
