package vesper

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
)

type chainIndex struct{}

func newTestMiddleware(name string) func(LambdaFunc) LambdaFunc {
	return func(next LambdaFunc) LambdaFunc {
		return func(ctx context.Context, i interface{}) (interface{}, error) {
			idx, _ := ctx.Value(chainIndex{}).(int)
			ctx = context.WithValue(ctx, chainIndex{}, idx+1)
			return next(context.WithValue(ctx, name, idx), i)
		}
	}
}

func assertMiddlewareCalled(ctx context.Context, t *testing.T, name string, expectedIndex int) {
	chainIndex, ok := ctx.Value(name).(int)
	if !ok {
		t.Errorf("could not find chain index")
	}
	assert.Equal(t, expectedIndex, chainIndex, "middleware %s was called out of order", name)
}

func TestInvalidHandlerSignature(t *testing.T) {
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
			rsp, err := v.buildHandler().Invoke(context.Background(), []byte("{}"))
			assert.Nil(t, rsp)
			assert.Error(t, err)
		})
	}
}

func TestValidHandlerSignatures(t *testing.T) {
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
			handlerFunc: func(events.SQSEvent) error {
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
			handlerFunc: func(events.SQSEvent) (int, error) {
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
			handlerFunc: func(context.Context, events.SQSEvent) error {
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
			handlerFunc: func(context.Context, events.SQSEvent) (int, error) {
				return 0, nil
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			v := New(test.handlerFunc)
			_, err := v.buildHandler().Invoke(context.Background(), []byte("{}"))
			assert.NoError(t, err)
		})
	}
}

func TestDisableAutoUnmarshal(t *testing.T) {
	handlerCalled := false
	h := func(ctx context.Context, in int) (int, error) {
		handlerCalled = true
		return 1, nil
	}
	m1 := func(next LambdaFunc) LambdaFunc {
		return func(ctx context.Context, in interface{}) (interface{}, error) {
			assert.Equal(t, []byte("1"), in, "expected input not to be unmarshaled into TIn type of int")
			return next(ctx, 1)
		}
	}
	v := New(h, m1)
	v.DisableAutoUnmarshal()
	_, err := v.buildHandler().Invoke(context.Background(), []byte("1"))
	assert.True(t, handlerCalled)
	assert.NoError(t, err)
}

func TestHandler(t *testing.T) {

	t.Run("no TOut should return string null in bytes", func(t *testing.T) {
		v := New(func(context.Context, interface{}) {})
		rsp, err := v.buildHandler().Invoke(context.Background(), []byte("{}"))
		assert.Equal(t, []byte("null"), rsp)
		assert.NoError(t, err)
	})

	t.Run("input to lambda handler from middlewares", func(t *testing.T) {
		t.Run("incompatible type", func(t *testing.T) {
			h := func(ctx context.Context, in int) (int, error) {
				t.Errorf("unexpected call to handler")
				return in, nil
			}
			m1 := func(next LambdaFunc) LambdaFunc {
				return func(ctx context.Context, in interface{}) (interface{}, error) {
					return next(ctx, "incompatible value")
				}
			}
			v := New(h, m1)
			v.DisableAutoUnmarshal()
			_, err := v.buildHandler().Invoke(context.Background(), []byte("{}"))
			assert.Error(t, err)
		})

		t.Run("nil input and TIn is value type", func(t *testing.T) {
			handlerCalled := false
			h := func(ctx context.Context, in int) (int, error) {
				handlerCalled = true
				assert.Equal(t, 0, in, "expected input to be zero value")
				return in, nil
			}
			m1 := func(next LambdaFunc) LambdaFunc {
				return func(ctx context.Context, in interface{}) (interface{}, error) {
					return next(ctx, nil)
				}
			}
			v := New(h, m1)
			v.DisableAutoUnmarshal()
			_, err := v.buildHandler().Invoke(context.Background(), []byte("{}"))
			assert.True(t, handlerCalled)
			assert.NoError(t, err)
		})

		t.Run("nil input and TIn is pointer type", func(t *testing.T) {
			handlerCalled := false
			h := func(ctx context.Context, in *int) (int, error) {
				handlerCalled = true
				assert.Nil(t, in, "expected input to be zero value")
				return 0, nil
			}
			m1 := func(next LambdaFunc) LambdaFunc {
				return func(ctx context.Context, in interface{}) (interface{}, error) {
					return next(ctx, nil)
				}
			}
			v := New(h, m1)
			v.DisableAutoUnmarshal()
			_, err := v.buildHandler().Invoke(context.Background(), []byte("{}"))
			assert.True(t, handlerCalled)
			assert.NoError(t, err)
		})
	})

	t.Run("no middlewares", func(t *testing.T) {
		t.Run("no error from handler", func(t *testing.T) {
			called := false
			v := New(func(context.Context, interface{}) (int, error) {
				called = true
				return 1, nil
			})
			rsp, err := v.buildHandler().Invoke(context.Background(), []byte("{}"))
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
			_, err := v.buildHandler().Invoke(context.Background(), []byte("{}"))
			assert.True(t, called)
			assert.Error(t, err)
		})
	})

	t.Run("middlewares provided", func(t *testing.T) {
		t.Run("no error", func(t *testing.T) {
			outerCtx := context.Background()
			handlerCalled := false
			h := func(ctx context.Context, i interface{}) (int, error) {
				handlerCalled = true
				outerCtx = ctx
				assert.Equal(t, 4, ctx.Value(chainIndex{}).(int))
				return 1, nil
			}
			v := New(h, newTestMiddleware("m1"))
			v.Use(newTestMiddleware("m2"), newTestMiddleware("m3"))
			v.Use(newTestMiddleware("m4"))
			_, err := v.buildHandler().Invoke(context.Background(), []byte("{}"))
			assertMiddlewareCalled(outerCtx, t, "m1", 0)
			assertMiddlewareCalled(outerCtx, t, "m2", 1)
			assertMiddlewareCalled(outerCtx, t, "m3", 2)
			assertMiddlewareCalled(outerCtx, t, "m4", 3)
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
			_, err := v.buildHandler().Invoke(context.Background(), []byte("{}"))
			assert.Error(t, err)
		})
		t.Run("error from handler", func(t *testing.T) {
			outerCtx := context.Background()
			handlerCalled := false
			h := func(ctx context.Context, i interface{}) (int, error) {
				outerCtx = ctx
				handlerCalled = true
				assert.Equal(t, 2, ctx.Value(chainIndex{}).(int))
				return 0, errors.New("something happened")
			}
			v := New(h, newTestMiddleware("m1"), newTestMiddleware("m2"))
			_, err := v.buildHandler().Invoke(outerCtx, []byte("{}"))
			assertMiddlewareCalled(outerCtx, t, "m1", 0)
			assertMiddlewareCalled(outerCtx, t, "m2", 1)
			assert.True(t, handlerCalled)
			assert.Error(t, err)
		})
	})
}
