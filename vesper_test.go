package vesper

import (
	"context"
	"fmt"
	"log"
	"testing"
)

// Typed interface for testing
type User struct {
	Username string
	Password string
}

// MyHandler implements the Lambda Handler interface
func MyHandler(ctx context.Context, u User) (interface{}, error) {
	fmt.Println("[actual handler]: Start. Username: ", u.Username)

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

// var TypedMiddleware = func(f LambdaFunc) LambdaFunc {
// 	// one time scope setup area for middleware

// 	return func(ctx context.Context, in interface{}) (interface{}, error) {
// 		fmt.Println("[dummyMiddleware] start: ", in)
// 		res, err := f(ctx, in)
// 		fmt.Println("[dummyMiddleware] END: ", res)

// 		return res, err
// 	}
// }

type warmupEvent struct {
	Event struct {
		Source string
	}
}

// WarmupMiddleware detects a warmup invocation event from the
// plugin "serverless-plugin-warmup", and returns early if found
//
// See https://www.npmjs.com/package/serverless-plugin-warmup for more
var WarmupMiddleware = func(f LambdaFunc) LambdaFunc {
	return func(ctx context.Context, in interface{}) (interface{}, error) {
		log.Println("[warmupMiddleware] start")
		var event warmupEvent

		if err := ExtractType(ctx, &event); err == nil {
			if event.Event.Source == "serverless-plugin-warmup" {
				log.Println("[warmupMiddleware] warmup event detected, exiting")
				return "warmup", nil
			}
		}

		res, err := f(ctx, in)
		log.Println("[warmupMiddleware] END: ", res)

		return res, err
	}
}

func Test_buildTypedMiddleware(t *testing.T) {
	// m := New(MyHandler, AuthMiddleware, DummyMiddleware)
	m := New(MyHandler, WarmupMiddleware)
	ctx := context.TODO()

	// The Lamda handler interface takes a context and a byte[]
	// i := []byte(`{"username":"matt", "password":"nottelling"}`)
	i := []byte(`{"event": {"source": "serverless-plugin-warmup"}}`)
	res, err := m.handler(ctx, i)

	fmt.Printf("res: %v, err: %v", res, err)
}
