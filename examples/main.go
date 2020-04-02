package main

import (
	"context"
	"fmt"
	"log"

	"github.com/mefellows/vesper"
	"github.com/mefellows/vesper/middleware"
)

type User struct {
	Username string
	Password string
}

// MyHandler implements the Lambda Handler interface
func MyHandler(ctx context.Context, u User) (interface{}, error) {
	log.Printf("[actual handler]: Have User %+v\n", u)

	return u.Username, nil
}

var correlationIDMiddleware = func(f vesper.LambdaFunc) vesper.LambdaFunc {
	return func(ctx context.Context, in interface{}) (interface{}, error) {
		log.Println("[correlationIdMiddleware] start")

		res, err := f(ctx, in)

		log.Println("[correlationIdMiddleware] END")

		return res, err
	}
}

var authMiddleware = func(f vesper.LambdaFunc) vesper.LambdaFunc {
	// one time scope setup area for middleware

	return func(ctx context.Context, in interface{}) (interface{}, error) {
		log.Println("[authMiddleware] start: ", in)
		user := in.(User)
		if user.Username == "fail" {
			error := map[string]string{
				"error": "unauthorised",
			}
			return error, fmt.Errorf("user %v is unauthorised", in)
		}

		res, err := f(ctx, in)
		log.Println("[authMiddleware] END: ", res)

		return res, err
	}
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
	m := vesper.New(MyHandler, middleware.WarmupMiddleware, correlationIDMiddleware, dummyMiddleware, authMiddleware)
	m.Start()
}
