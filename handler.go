package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"reflect"

	"github.com/aws/aws-lambda-go/lambda"
)

// TODO: 1. Support passing in functions, instead of structs
//       2. Support modification of request/response (e.g. pass through objects?)
//       3. ...WithContext handler wrapper? -> dah, already in the interface!! - DONE
//       4. Logging
//       5. Confirm what types Lambda accepts in its interface - DONE (not primitives!)

// LambdaFunc is (long-form) of the Lambda handler interface
// https://github.com/aws/aws-lambda-go/blob/master/lambda/entry.go#L37-L49
type LambdaFunc func(context.Context, interface{}) (interface{}, error)

// Handler is the interface to serve as Middleware in Middy
type Handler interface {
	Handle(ctx context.Context, in interface{}, next LambdaFunc) (interface{}, error)
}

// HandlerFunc is an adapter to allow normal functions to act as a Handler
type HandlerFunc func(ctx context.Context, in interface{}, next LambdaFunc) (interface{}, error)

// Handle implements the Handle interface for a first-class Handler
func (h HandlerFunc) Handle(ctx context.Context, in interface{}, next LambdaFunc) (interface{}, error) {
	return h(ctx, in, next)
}

// Middleware is a one-way linked list of middleware Handlers
// This is negroni-style middleware. The linked-list gives us the ability to modify
// the order etc. a bit more nicely than the pure functional style middleware
type middleware struct {
	handler Handler
	next    *middleware
}

func (m middleware) Handle(ctx context.Context, in interface{}) (interface{}, error) {
	return m.handler.Handle(ctx, in, m.next.Handle)
}

// Middy is a middleware adapter for Lambda Functions
// Middleware are evaluated in the order they are added to the stack
type Middy struct {
	handlers   []Handler
	middleware middleware
}

// New creates a new Middy given a set of Handlers
func New(l interface{}, handlers ...Handler) *Middy {
	handlers = append(handlers, wrapUserHandler(l))

	return &Middy{
		handlers:   handlers,
		middleware: build(handlers),
	}
}

func (m *Middy) with(f LambdaFunc) *Middy {
	return m
}

func (m *Middy) WithType(i interface{}) *Middy {
	return m
}

// Start is a convienence function run the lambda handler
func (m *Middy) Start(ctx context.Context, in interface{}) {
	lambda.Start(m.Handle())
}

// Handle yields the Lambda compatible function handler
// with all middleware applied
func (m *Middy) Handle() LambdaFunc {
	return m.middleware.Handle
}

func build(handlers []Handler) middleware {
	var next middleware

	if len(handlers) == 0 {
		fmt.Println("Adding void middleware - this means no middleware provided and is a noop")
		return voidMiddleware()
	} else if len(handlers) > 1 {
		next = build(handlers[1:])
	} else {
		fmt.Printf("Adding void middleware - this will be linked to the wrapped handler (which will never call it)")
		next = voidMiddleware()
	}

	return middleware{handlers[0], &next}
}

func typeMiddleware() middleware {
	return middleware{
		HandlerFunc(func(ctx context.Context, in interface{}, next LambdaFunc) (interface{}, error) {
			fmt.Println("Type middleware")
			var args []reflect.Value
			args = append(args, reflect.ValueOf(ctx))
			args = append(args, reflect.ValueOf(in))
			var out interface{}

			if err := json.Unmarshal([]byte(`{"foo":"cached"}`), &out); err != nil {
				return nil, err
			}

			return out, nil
		}),
		&middleware{},
	}
}

func voidMiddleware() middleware {
	return middleware{
		HandlerFunc(func(ctx context.Context, in interface{}, next LambdaFunc) (interface{}, error) {
			log.Fatal("Void middleware - should never be executed!")
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
		// TODO: error here?
		fmt.Printf("handler kind %s is not %s", handlerType.Kind(), reflect.Func)
	}

	// Wrap the handler into a generic Handler type
	// This handler _does not_ execute the next middleware (which should be the voidMiddleware)
	return HandlerFunc(func(ctx context.Context, in interface{}, next LambdaFunc) (interface{}, error) {
		fmt.Println("Wrapped lambda handler")
		var args []reflect.Value
		args = append(args, reflect.ValueOf(ctx))
		args = append(args, reflect.ValueOf(in))

		// // convert return values into (interface{}, error)
		response := handler.Call(args)
		// args = append(args)
		// response := handler.Call(reflect.ValueOf(ctx), reflect.ValueOf(in))

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
