package vesper

import (
	"context"
	"reflect"
	"testing"

	"github.com/aws/aws-lambda-go/events"
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
	_, _ = v.buildHandler().Invoke(context.Background(), in)
	assert.True(t, called)
}

func TestTInFromContext(t *testing.T) {
	t.Run("no TIn in handler signature", func(t *testing.T) {
		called := false
		v := New(func(ctx context.Context) error {
			_, ok := TInFromContext(ctx)
			assert.False(t, ok)
			called = true
			return nil
		})
		_, _ = v.buildHandler().Invoke(context.Background(), []byte("{}"))
		assert.True(t, called)
	})

	t.Run("TIn in handler signature", func(t *testing.T) {
		called := false
		handler := func(ctx context.Context, evt events.SQSEvent) error {
			tIn, ok := TInFromContext(ctx)
			assert.True(t, ok)
			assert.Equal(t, reflect.TypeOf(events.SQSEvent{}), tIn)
			called = true
			return nil
		}
		v := New(handler)
		_, _ = v.buildHandler().Invoke(context.Background(), []byte("{}"))
		assert.True(t, called)
	})
}
