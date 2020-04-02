package vesper

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPayloadFromContext(t *testing.T) {
	in := []byte(`
{
	"message": "hello world"
}
`)
	called := false
	v := New(func(ctx context.Context, i interface{}) (int, error) {
		payload, ok := PayloadFromContext(ctx)
		assert.True(t, ok)
		assert.Equal(t, in, payload)
		called = true
		return 1, nil
	})
	_, _ = v.handler.Invoke(context.Background(), in)
	assert.True(t, called)
}

func TestHandlerSignatureFromContext(t *testing.T) {
	type user struct {
		Name string
	}
	called := false
	v := New(func(ctx context.Context, u user) error {
		sig, ok := HandlerSignatureFromContext(ctx)
		assert.True(t, ok)
		assert.IsType(t, reflect.TypeOf(user{}), sig.In(1))
		assert.Equal(t, "func(context.Context, vesper.user) error", sig.String())
		called = true
		return nil
	})
	_, _ = v.handler.Invoke(context.Background(), []byte("{}"))
	assert.True(t, called)
}
