package main

import (
	"context"
	"fmt"
)

// MyHandler implements the Lambda Handler interface
func MyHandler(ctx context.Context, o foo) (interface{}, error) {
	fmt.Println("Received body: ", ctx, o.Foo)

	return "ok", nil
}

// MyHandlerMiddleware is a test middleware
type MyHandlerMiddleware struct{}

// Start implements the handler interface.Start
// TODO: allow this to be a function so we can simplify the interface
func (m MyHandlerMiddleware) Start(ctx context.Context, in interface{}, next LambdaFunc) (interface{}, error) {
	fmt.Println("MyHandlerMiddleware start: ", ctx, in)
	fmt.Println("Next Handler: ", next, in)

	return next(ctx, in)
}

func main() {
	// m := New(MyHandlerMiddleware{})
	// m.Start(ctx context.Context, in interface{})
}
