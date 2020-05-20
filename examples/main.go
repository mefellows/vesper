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
	log.Printf("\n\n[userHandler]: Have User %v \n\n", u)
	var err error

	if u.Username == "" || u.Password == "" {
		err = errors.New("unauthorised")
	}

	return Response{
		Authenticated: true,
	}, err
}

// LoginHandler implements the Lambda Handler interface
func LoginHandler2(ctx context.Context, u User) (interface{}, error) {
	log.Println("[userHandler] checking request type for warmup event or API call")
	log.Println("[userHandler] Validating API request")
	log.Println("[userHandler] Extracting correlation ID for event")
	log.Println("[userHandler] Authenticating request")
	log.Println("[userHandler] Executing User API Call")
	log.Println("[userHandler] validating output response body")
	log.Println("[userHandler] setting correlation ID property")

	return nil, nil
}

var namedMiddleware = func(name string) vesper.Middleware {
	return func(f vesper.LambdaFunc) vesper.LambdaFunc {
		return func(ctx context.Context, in interface{}) (interface{}, error) {
			log.Printf("[%s] START: %+v", name)

			res, err := f(ctx, in)

			log.Printf("[%s] END %+v", name)

			return res, err
		}
	}
}

var authMiddleware = func(f vesper.LambdaFunc) vesper.LambdaFunc {
	// one time scope setup area for middleware

	return func(ctx context.Context, in interface{}) (interface{}, error) {
		log.Println("[authMiddleware] START: ", in)
		user := in.(User)
		if user.Username == "" || user.Password == "" {
			error := map[string]string{
				"error": "unauthorised",
			}
			log.Println("[authMiddleware] user is unauthorised, short circuiting request: ", in)
			return error, fmt.Errorf("user %v is unauthorised", user.Username)
		}

		res, err := f(ctx, in)
		log.Printf("[authMiddleware] END: %+v \n", in)

		return res, err
	}
}

func main() {
	vesper.Logger(log.New(os.Stdout, "", log.LstdFlags))
	// m := vesper.New(LoginHandler2)
	m := vesper.New(LoginHandler, namedMiddleware("requestValidationMiddleware"), namedMiddleware("correlationIdMiddleware"), authMiddleware)
	m.Start()
}
