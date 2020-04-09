package vesper

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
)

func TestSQSParser(t *testing.T) {
	type user struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	t.Run("no unmarshaler", func(t *testing.T) {
		nextFunc := func(ctx context.Context, in interface{}) (interface{}, error) {
			t.Errorf("unexpected call to next func")
			return nil, nil
		}
		middleware := SQSParserMiddleware(nil)(nextFunc)
		_, err := middleware(context.Background(), []byte("{}"))
		assert.Error(t, err)
	})

	t.Run("no TIn", func(t *testing.T) {
		called := false
		nextFunc := func(ctx context.Context, in interface{}) (interface{}, error) {
			called = true
			return nil, nil
		}
		middleware := SQSParserMiddleware(json.Unmarshal)(nextFunc)
		payload := []byte("{}")
		ctx := context.Background()
		ctx = context.WithValue(ctx, ctxKeyPayload, payload)
		_, err := middleware(ctx, payload)
		assert.NoError(t, err)
		assert.True(t, called)
	})

	t.Run("TIn is not a slice", func(t *testing.T) {
		nextFunc := func(ctx context.Context, in interface{}) (interface{}, error) {
			t.Errorf("unexpected call to next func")
			return nil, nil
		}
		middleware := SQSParserMiddleware(json.Unmarshal)(nextFunc)
		payload := []byte("{}")
		ctx := context.Background()
		ctx = context.WithValue(ctx, ctxKeyTIn, reflect.TypeOf(user{}))
		ctx = context.WithValue(ctx, ctxKeyPayload, payload)
		_, err := middleware(ctx, payload)
		assert.Error(t, err)
	})

	t.Run("TIn is of type events.SQSEvent", func(t *testing.T) {
		nextFunc := func(ctx context.Context, in interface{}) (interface{}, error) {
			t.Errorf("unexpected call to next func")
			return nil, nil
		}
		middleware := SQSParserMiddleware(json.Unmarshal)(nextFunc)
		payload := []byte("{}")
		ctx := context.Background()
		ctx = context.WithValue(ctx, ctxKeyTIn, reflect.TypeOf(events.SQSEvent{}))
		ctx = context.WithValue(ctx, ctxKeyPayload, payload)
		_, err := middleware(ctx, payload)
		assert.Error(t, err)
	})

	t.Run("payload cannot be unmarshalled to SQSEvent", func(t *testing.T) {
		nextFunc := func(ctx context.Context, in interface{}) (interface{}, error) {
			t.Errorf("unexpected call to next func")
			return nil, nil
		}
		middleware := SQSParserMiddleware(json.Unmarshal)(nextFunc)
		payload := []byte("not an SQS event")
		ctx := context.Background()
		ctx = context.WithValue(ctx, ctxKeyTIn, reflect.TypeOf([]user{}))
		ctx = context.WithValue(ctx, ctxKeyPayload, payload)
		_, err := middleware(ctx, payload)
		assert.Error(t, err)
	})

	t.Run("TIn is a slice of interface types", func(t *testing.T) {
		request := `
{
  "Records": [
    {
      "messageId": "19dd0b57-b21e-4ac1-bd88-01bbb068cb78",
      "body": "{\"name\": \"myuser\", \"age\": 20}"
    },
    {
      "messageId": "fa9c517a-1a5d-4109-b689-081ceee6edbb",
      "body": "{\"name\": \"another user\", \"age\": 18}"
    }
  ]
}
`
		nextFunc := func(ctx context.Context, in interface{}) (interface{}, error) {
			assert.IsType(t, []user{}, in)
			assert.Len(t, in, 2)
			return nil, nil
		}
		middleware := SQSParserMiddleware(json.Unmarshal)(nextFunc)
		payload := []byte(request)
		ctx := context.Background()
		ctx = context.WithValue(ctx, ctxKeyTIn, reflect.TypeOf([]user{}))
		ctx = context.WithValue(ctx, ctxKeyPayload, payload)
		_, err := middleware(ctx, payload)
		assert.NoError(t, err)
	})

	t.Run("payload cannot be unmarshalled to SQSEvent", func(t *testing.T) {
		request := `
{
  "Records": [
    {
      "messageId": "19dd0b57-b21e-4ac1-bd88-01bbb068cb78",
      "body": "this is not a valid user"
    },
    {
      "messageId": "fa9c517a-1a5d-4109-b689-081ceee6edbb",
      "body": "{\"name\": \"another user\", \"age\": 18}"
    }
  ]
}
`
		nextFunc := func(ctx context.Context, in interface{}) (interface{}, error) {
			t.Errorf("unexpected call to next func")
			return nil, nil
		}
		middleware := SQSParserMiddleware(json.Unmarshal)(nextFunc)
		payload := []byte(request)
		ctx := context.Background()
		ctx = context.WithValue(ctx, ctxKeyTIn, reflect.TypeOf([]user{}))
		ctx = context.WithValue(ctx, ctxKeyPayload, payload)
		_, err := middleware(ctx, payload)
		assert.Error(t, err)
	})

	t.Run("happy path", func(t *testing.T) {
		tests := []struct {
			name          string
			payload       string
			handlerCalled bool
			expectedTIn   []user
			wantErr       bool
		}{
			{
				name:        "empty records",
				payload:     `{"Records": []}`,
				expectedTIn: []user{},
			},
			{
				name: "multiple user records in payload",
				payload: `
{
  "Records": [
    {
      "messageId": "19dd0b57-b21e-4ac1-bd88-01bbb068cb78",
      "body": "{\"name\": \"myuser\", \"age\": 20}"
    },
    {
      "messageId": "fa9c517a-1a5d-4109-b689-081ceee6edbb",
      "body": "{\"name\": \"another user\", \"age\": 18}"
    }
  ]
}
`,
				handlerCalled: true,
				expectedTIn: []user{
					{Name: "myuser", Age: 20},
					{Name: "another user", Age: 18},
				},
				wantErr: false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				nextFuncCalled := false
				nextFunc := func(ctx context.Context, in interface{}) (interface{}, error) {
					nextFuncCalled = true
					assert.Equal(t, tt.expectedTIn, in)
					return nil, nil
				}
				middleware := SQSParserMiddleware(json.Unmarshal)(nextFunc)
				payload := []byte(tt.payload)
				ctx := context.Background()
				ctx = context.WithValue(ctx, ctxKeyTIn, reflect.TypeOf([]user{}))
				ctx = context.WithValue(ctx, ctxKeyPayload, payload)
				_, err := middleware(ctx, payload)
				assert.True(t, nextFuncCalled)
				assert.NoError(t, err)
			})
		}
	})
}
