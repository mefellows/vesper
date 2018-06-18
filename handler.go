package main

import (
	"context"
	"fmt"
	"reflect"
)

// TODO: 1. Support passing in functions, instead of structs
//       2. Support modification of request/response (e.g. pass through objects?)
//       3. Refactor "Start" to something like "Handle"

// LambdaFunc is (long-form) of the Lambda handler interface
// https://github.com/aws/aws-lambda-go/blob/master/lambda/entry.go#L37-L49
type LambdaFunc func(context.Context, interface{}) (interface{}, error)

// Handler is the interface to serve as Middleware in Middy
type Handler interface {
	Start(ctx context.Context, in interface{}, next LambdaFunc) (interface{}, error)
}

// HandlerFunc is an adapter to allow normal functions to act as a Handler
type HandlerFunc func(ctx context.Context, in interface{}, next LambdaFunc) (interface{}, error)

// Start implements the Start interface for a first-class Handler
func (h HandlerFunc) Start(ctx context.Context, in interface{}, next LambdaFunc) (interface{}, error) {
	return h(ctx, in, next)
}

// Middleware is a one-way linked list of middleware Handlers
type middleware struct {
	handler Handler
	next    *middleware
}

func (m middleware) Start(ctx context.Context, in interface{}) (interface{}, error) {
	return m.handler.Start(ctx, in, m.next.Start)
}

// Middy is a middleware adapter for Lambda Functions
// Middleware are evaluated in the order they are added to the stack
type Middy struct {
	lambda     LambdaFunc
	handlers   []Handler
	middleware middleware
	// lambdaStartHandler func(interface{}) // Handler to call
}

// New creates a new Middy given a set of Handlers
func New(l interface{}, handlers ...Handler) *Middy {
	handlers = append(handlers, wrapUserHandler(l))

	return &Middy{
		handlers:   handlers,
		middleware: build(handlers),
	}
}

// Start is the Lambda compatible interface for Middy
func (m *Middy) Start(ctx context.Context, in interface{}) (interface{}, error) {
	return m.middleware.Start(ctx, in)
}

func build(handlers []Handler) middleware {
	var next middleware

	if len(handlers) == 0 {
		fmt.Println("Adding void middleware")
		return voidMiddleware()
	} else if len(handlers) > 1 {
		next = build(handlers[1:])
	} else {
		fmt.Println("Adding void middleware")
		next = voidMiddleware()
	}

	return middleware{handlers[0], &next}
}

// TODO: should this be the wrapped initial handler?
func voidMiddleware() middleware {
	return middleware{
		HandlerFunc(func(ctx context.Context, in interface{}, next LambdaFunc) (interface{}, error) {
			fmt.Println("Void middleware!")
			return nil, nil
		}),
		&middleware{},
	}
}

func wrapLambda(l LambdaFunc) Handler {
	return HandlerFunc(func(ctx context.Context, in interface{}, next LambdaFunc) (interface{}, error) {
		fmt.Println("Original handler")
		res, err := l(ctx, in)
		fmt.Printf("Response from original handler: %v, err: %v", res, err)
		return next(ctx, in)
	})
}

func wrapUserHandler(l interface{}) Handler {
	// Wrap the passed in handler, as it may/should have proper argument types
	// e.g. https://github.com/aws/aws-lambda-go/blob/master/lambda/handler.go#L75
	handler := reflect.ValueOf(l)
	handlerType := reflect.TypeOf(l)
	if handlerType.Kind() != reflect.Func {
		fmt.Printf("handler kind %s is not %s", handlerType.Kind(), reflect.Func)
		//return errorHandler(fmt.Errorf("handler kind %s is not %s", handlerType.Kind(), reflect.Func))
	}

	// Wrap the handler into a generic Handler type
	return HandlerFunc(func(ctx context.Context, in interface{}, next LambdaFunc) (interface{}, error) {
		fmt.Println("Wrapped lambda handler")
		var args []reflect.Value
		args = append(args, reflect.ValueOf(ctx))
		args = append(args, reflect.ValueOf(in))

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

		fmt.Println("Wrapped lambda handler. Response from orig handler:", val, err)

		return val, nil
	})
}
