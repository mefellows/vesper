package main

import (
	"context"
	"fmt"
)

// Typed interface for testing
type User struct {
	Username string
	Password string
}

// MyHandler implements the Lambda Handler interface
func MyHandler(ctx context.Context, u User) (interface{}, error) {
	fmt.Println("[actual handler]: Starte. Username: ", u.Username)

	return u.Username, nil
}

var AuthMiddleware = func(f LambdaFunc) LambdaFunc {
	// one time scope setup area for middleware

	return func(ctx context.Context, in interface{}) (interface{}, error) {
		fmt.Println("[authMiddleware] start: ", in)
		user := in.(User)
		if user.Username == "fail" {
			error := map[string]string{
				"error": "unauthorised",
			}
			return error, fmt.Errorf("user %v is unauthorised", in)
		}

		res, err := f(ctx, in)
		fmt.Println("[authMiddleware] END: ", res)

		return res, err
	}
}

var DummyMiddleware = func(f LambdaFunc) LambdaFunc {
	// one time scope setup area for middleware

	return func(ctx context.Context, in interface{}) (interface{}, error) {
		fmt.Println("[dummyMiddleware] start: ", in)
		res, err := f(ctx, in)
		fmt.Println("[dummyMiddleware] END: ", res)

		return res, err
	}
}

func main() {
	m := New(MyHandler, AuthMiddleware, DummyMiddleware)
	m.Start()
}
