package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/mefellows/vesper"
)

type User struct {
	Username string
	Password string
}

type Response struct {
	Authenticated bool `json:"authenticated"`
}

// LoginHandler implements the Lambda Handler interface
func LoginHandler(ctx context.Context, u User) (Response, error) {
	log.Printf("[actual handler]: Have User %v", u)
	var err error

	if u.Username == "" || u.Password == "" {
		err = errors.New("unauthorised")
	}

	return Response{
		Authenticated: true,
	}, err
}

var correlationIDMiddleware = func(f vesper.LambdaFunc) vesper.LambdaFunc {
	return func(ctx context.Context, in interface{}) (interface{}, error) {
		log.Println("[correlationIdMiddleware] START:", in)

		res, err := f(ctx, in)

		log.Println("[correlationIdMiddleware] END ", in)

		return res, err
	}
}

var authMiddleware = func(f vesper.LambdaFunc) vesper.LambdaFunc {
	// one time scope setup area for middleware

	return func(ctx context.Context, in interface{}) (interface{}, error) {
		log.Println("[authMiddleware] START: ", in)
		user := in.(User)
		if user.Username == "fail" {
			error := map[string]string{
				"error": "unauthorised",
			}
			return error, fmt.Errorf("user %v is unauthorised", in)
		}

		res, err := f(ctx, in)
		log.Printf("[authMiddleware] END: %+v \n", in)

		return res, err
	}
}

var dummyMiddleware = func(f vesper.LambdaFunc) vesper.LambdaFunc {
	// one time scope setup area for middleware

	return func(ctx context.Context, in interface{}) (interface{}, error) {
		log.Println("[dummyMiddleware] START:")
		res, err := f(ctx, in)
		log.Println("[dummyMiddleware] END:", in)

		return res, err
	}
}

func main() {
	// m := vesper.New(LoginHandler)
	m := vesper.New(LoginHandler, vesper.WarmupMiddleware, correlationIDMiddleware, dummyMiddleware, authMiddleware)
	vesper.Logger(log.New(os.Stdout, "", log.LstdFlags))
	m.Start()
}
