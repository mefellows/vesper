package main

import (
	"context"
	"log"

	"github.com/mefellows/vesper"
)

// MyHandler implements the Lambda Handler interface
// NOTE: this handlerhttps://github.com/mefellows/vesper
func MyHandler(ctx context.Context, o interface{}) (interface{}, error) {
	log.Println("[actual handler]: Start")

	return o, nil
}

var dummyMiddleware = func(f vesper.LambdaFunc) vesper.LambdaFunc {
	// one time scope setup area for middleware

	return func(ctx context.Context, in interface{}) (interface{}, error) {
		log.Println("[dummyMiddleware] start: ", in)
		res, err := f(ctx, in)
		log.Println("[dummyMiddleware] END: ", res)

		return res, err
	}
}

func main() {
	m := vesper.New(MyHandler, dummyMiddleware)
	m.Start()
}
