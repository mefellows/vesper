package main

import (
	"context"
	"fmt"
	"reflect"
	"testing"
)

type foo struct {
	Foo string
}

var fakeHandler = func(handlerInput interface{}) {
	fmt.Println("LAMBDA HANDLER: executing original handler")
	var args []reflect.Value
	ctx := context.TODO()
	i := foo{
		Foo: "bar",
	}
	args = append(args, reflect.ValueOf(ctx))
	args = append(args, reflect.ValueOf(i))
	handler := reflect.ValueOf(handlerInput)

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

	fmt.Println("LAMBDA HANDLER: Response after middleware:", val, err)
}

// MyHandler implements the Lambda Handler interface
func testHandler(ctx context.Context, o foo) (interface{}, error) {
	fmt.Println("ACTUAL HANDLER: Received body: ", ctx, o.Foo)

	return "ok", nil
}

// testHandlerMiddleware is a test middleware
type testHandlerMiddleware struct{}

// Start implements the handler interface.Start
// TODO: allow this to be a function so we can simplify the interface
func (m testHandlerMiddleware) Start(ctx context.Context, in interface{}, next LambdaFunc) (interface{}, error) {
	fmt.Println("testHandlerMiddleware START: ", ctx, in)

	newIn := foo{
		Foo: "not bar!",
	}

	return next(ctx, newIn)
}

// testHandlerMiddleware is a test middleware
type yetAnotherMiddleware struct{}

// Start implements the handler interface.Start
// TODO: allow this to be a function so we can simplify the interface
func (m yetAnotherMiddleware) Start(ctx context.Context, in interface{}, next LambdaFunc) (interface{}, error) {
	fmt.Println("Yet another middleware start: ", ctx, in)

	res, err := next(ctx, in)
	fmt.Println("Yet another middleware END: ", ctx, in)

	return res, err
}

func Test_middleware(t *testing.T) {
	fmt.Println("Testing")
	ctx := context.TODO()
	i := foo{
		Foo: "bar",
	}
	//var input interface{}
	//input = map[string]string{
	//	"foo": "bar",
	//}

	m := New(testHandler, yetAnotherMiddleware{}, testHandlerMiddleware{})
	res, err := m.Start(ctx, i)
	fmt.Printf("res: %v, err: %v", res, err)

}
