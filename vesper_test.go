package vesper

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandler(t *testing.T) {

	t.Run("invalid handler", func(t *testing.T) {
		tests := []struct {
			name        string
			handlerFunc interface{}
		}{
			{
				name: "more than 2 return values",
				handlerFunc: func(context.Context, interface{}) (int, int, error) {
					return 1, 2, nil
				},
			},
			{
				name: "2 return values but error is not second param",
				handlerFunc: func(context.Context, interface{}) (int, int) {
					return 1, 2
				},
			},
			{
				name: "1 return value which is not error type",
				handlerFunc: func(context.Context, interface{}) int {
					return 1
				},
			},
			{
				name:        "nil handler",
				handlerFunc: nil,
			},
			{
				name:        "non func type",
				handlerFunc: "hello world!",
			},
			{
				name: "more than 2 parameters",
				handlerFunc: func(context.Context, interface{}, interface{}) error {
					return nil
				},
			},
			{
				name: "2 parameters but context is not first param",
				handlerFunc: func(int, interface{}) error {
					return nil
				},
			},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				v := New(test.handlerFunc)
				rsp, err := v.handler.Invoke(context.Background(), []byte("{}"))
				assert.Nil(t, rsp)
				assert.Error(t, err)
			})
		}
	})

	t.Run("valid handler", func(t *testing.T) {
		tests := []struct {
			name        string
			handlerFunc interface{}
		}{
			{
				name:        "func()",
				handlerFunc: func() {},
			},
			{
				name: "func()error",
				handlerFunc: func() error {
					return nil
				},
			},
			{
				name: "func(TIn)error",
				handlerFunc: func(int) error {
					return nil
				},
			},
			{
				name: "func()(TOut,error)",
				handlerFunc: func() (int, error) {
					return 0, nil
				},
			},
			{
				name: "func(TIn)(TOut,error)",
				handlerFunc: func(int) (int, error) {
					return 0, nil
				},
			},
			{
				name: "func(context.Context)error",
				handlerFunc: func(context.Context) error {
					return nil
				},
			},
			{
				name: "func(context.Context,TIn)error",
				handlerFunc: func(context.Context, int) error {
					return nil
				},
			},
			{
				name: "func(context.Context)(TOut,error)",
				handlerFunc: func(context.Context) (int, error) {
					return 0, nil
				},
			},
			{
				name: "func(context.Context,TIn)(TOut,error)",
				handlerFunc: func(context.Context, int) (int, error) {
					return 0, nil
				},
			},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				v := New(test.handlerFunc)
				_, err := v.handler.Invoke(context.Background(), []byte("0"))
				assert.NoError(t, err)
			})
		}
	})

	t.Run("no return value", func(t *testing.T) {
		v := New(func(context.Context, interface{}) {})
		rsp, err := v.handler.Invoke(context.Background(), []byte("{}"))
		assert.Equal(t, []byte("null"), rsp)
		assert.NoError(t, err)
	})

	t.Run("no middlewares", func(t *testing.T) {
		t.Run("no error from handler", func(t *testing.T) {
			called := false
			v := New(func(context.Context, interface{}) (int, error) {
				called = true
				return 1, nil
			})
			rsp, err := v.handler.Invoke(context.Background(), []byte("{}"))
			assert.Equal(t, []byte("1"), rsp)
			assert.True(t, called)
			assert.NoError(t, err)
		})
		t.Run("error from handler", func(t *testing.T) {
			called := false
			v := New(func(context.Context, interface{}) (int, error) {
				called = true
				return 1, errors.New("something happened")
			})
			_, err := v.handler.Invoke(context.Background(), []byte("{}"))
			assert.True(t, called)
			assert.Error(t, err)
		})
	})

	t.Run("middlewares provided", func(t *testing.T) {

		newTestMiddleware := func(name string) func(LambdaFunc) LambdaFunc {
			return func(next LambdaFunc) LambdaFunc {
				return func(ctx context.Context, i interface{}) (interface{}, error) {
					return next(context.WithValue(ctx, name, name), i)
				}
			}
		}

		checkMiddlewareCalled := func(ctx context.Context, name string) bool {
			return ctx.Value(name) != nil
		}

		t.Run("no error", func(t *testing.T) {
			outerCtx := context.Background()
			handlerCalled := false
			h := func(ctx context.Context, i interface{}) (int, error) {
				handlerCalled = true
				outerCtx = ctx
				return 1, nil
			}
			v := New(h, newTestMiddleware("m1"), newTestMiddleware("m2"))
			rsp, err := v.handler.Invoke(outerCtx, []byte("{}"))
			assert.Equal(t, []byte("1"), rsp)
			assert.True(t, checkMiddlewareCalled(outerCtx, "m1"))
			assert.True(t, checkMiddlewareCalled(outerCtx, "m2"))
			assert.True(t, handlerCalled)
			assert.NoError(t, err)
		})
		t.Run("error from middleware", func(t *testing.T) {
			m1 := func(next LambdaFunc) LambdaFunc {
				return func(ctx context.Context, in interface{}) (interface{}, error) {
					return nil, errors.New("something happened")
				}
			}
			m2 := func(next LambdaFunc) LambdaFunc {
				return func(ctx context.Context, in interface{}) (interface{}, error) {
					t.Error("second middleware should not have been called")
					return next(ctx, in)
				}
			}
			h := func(context.Context, interface{}) (int, error) {
				t.Error("handler should not have been called")
				return 1, nil
			}
			v := New(h, m1, m2)
			_, err := v.handler.Invoke(context.Background(), []byte("{}"))
			assert.Error(t, err)
		})
		t.Run("error from handler", func(t *testing.T) {
			outerCtx := context.Background()
			handlerCalled := false
			h := func(ctx context.Context, i interface{}) (int, error) {
				outerCtx = ctx
				handlerCalled = true
				return 0, errors.New("something happened")
			}
			v := New(h, newTestMiddleware("m1"), newTestMiddleware("m2"))
			_, err := v.handler.Invoke(outerCtx, []byte("{}"))
			assert.True(t, checkMiddlewareCalled(outerCtx, "m1"))
			assert.True(t, checkMiddlewareCalled(outerCtx, "m2"))
			assert.True(t, handlerCalled)
			assert.Error(t, err)
		})
	})
}
