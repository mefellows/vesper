package vesper

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParserMiddleware(t *testing.T) {

	t.Run("no unmarshaler", func(t *testing.T) {
		nextFunc := func(ctx context.Context, in interface{}) (interface{}, error) {
			t.Errorf("unexpected call to next func")
			return nil, nil
		}
		middleware := ParserMiddleware(nil)(nextFunc)
		_, err := middleware(context.Background(), []byte("{}"))
		assert.Error(t, err)
	})

	t.Run("no TIn", func(t *testing.T) {
		called := false
		nextFunc := func(ctx context.Context, in interface{}) (interface{}, error) {
			called = true
			return nil, nil
		}
		middleware := ParserMiddleware(json.Unmarshal)(nextFunc)
		payload := []byte("{}")
		ctx := context.Background()
		ctx = context.WithValue(ctx, ctxKeyPayload, payload)
		_, err := middleware(ctx, payload)
		assert.NoError(t, err)
		assert.True(t, called)
	})

	t.Run("TIn is a primitive", func(t *testing.T) {
		called := false
		nextFunc := func(ctx context.Context, in interface{}) (interface{}, error) {
			called = true
			assert.Equal(t, 123, in)
			return nil, nil
		}
		middleware := ParserMiddleware(json.Unmarshal)(nextFunc)
		payload := []byte("123")
		ctx := context.Background()
		ctx = context.WithValue(ctx, ctxKeyTIn, reflect.TypeOf(int(0)))
		ctx = context.WithValue(ctx, ctxKeyPayload, payload)
		_, err := middleware(ctx, payload)
		assert.NoError(t, err)
		assert.True(t, called)
	})

	t.Run("TIn is an object", func(t *testing.T) {
		called := false
		type req struct {
			Message string `json:"message"`
		}
		nextFunc := func(ctx context.Context, in interface{}) (interface{}, error) {
			called = true
			assert.Equal(t, req{Message: "hello world"}, in)
			return nil, nil
		}
		middleware := ParserMiddleware(json.Unmarshal)(nextFunc)
		payload := []byte(`{"message": "hello world"}`)
		ctx := context.Background()
		ctx = context.WithValue(ctx, ctxKeyTIn, reflect.TypeOf(req{}))
		ctx = context.WithValue(ctx, ctxKeyPayload, payload)
		_, err := middleware(ctx, payload)
		assert.NoError(t, err)
		assert.True(t, called)
	})
}
