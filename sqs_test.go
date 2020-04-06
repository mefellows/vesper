package vesper

import (
	"context"
	"testing"

	"github.com/mefellows/vesper/encoding"

	"github.com/aws/aws-lambda-go/events"

	"github.com/stretchr/testify/assert"
)

func TestSQSParser(t *testing.T) {
	type user struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	t.Run("no unmarshaler", func(t *testing.T) {
		v := New(func(context.Context) {}, SQSParser(nil))
		_, err := v.handler.Invoke(context.Background(), []byte("{}"))
		assert.Error(t, err)
	})

	t.Run("no TIn", func(t *testing.T) {
		v := New(func(context.Context) {}, SQSParser(encoding.JSONUnmarshal))
		_, err := v.handler.Invoke(context.Background(), []byte("{}"))
		assert.Error(t, err)
	})

	t.Run("TIn is not a slice", func(t *testing.T) {
		v := New(func(context.Context, user) {}, SQSParser(encoding.JSONUnmarshal))
		_, err := v.handler.Invoke(context.Background(), []byte("{}"))
		assert.Error(t, err)
	})

	t.Run("TIn is of type events.SQSEvent", func(t *testing.T) {
		v := New(func(context.Context, events.SQSEvent) {}, SQSParser(encoding.JSONUnmarshal))
		_, err := v.handler.Invoke(context.Background(), []byte("{}"))
		assert.Error(t, err)
	})

	t.Run("payload cannot be unmarshalled to SQSEvent", func(t *testing.T) {
		v := New(func(context.Context, []user) {}, SQSParser(encoding.JSONUnmarshal))
		_, err := v.handler.Invoke(context.Background(), []byte("not an SQS event"))
		assert.Error(t, err)
	})

	t.Run("TIn is a slice of interface types", func(t *testing.T) {
		payload := `
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
		v := New(func(ctx context.Context, in []interface{}) {
			assert.Len(t, in, 2)
		}, SQSParser(encoding.JSONUnmarshal))
		_, err := v.handler.Invoke(context.Background(), []byte(payload))
		assert.NoError(t, err)
	})

	t.Run("payload cannot be unmarshalled to SQSEvent", func(t *testing.T) {
		payload := `
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
		v := New(func(context.Context, []user) {}, SQSParser(encoding.JSONUnmarshal))
		_, err := v.handler.Invoke(context.Background(), []byte(payload))
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
				called := false
				handler := func(ctx context.Context, users []user) (interface{}, error) {
					called = true
					assert.Equal(t, tt.expectedTIn, users)
					return nil, nil
				}
				v := New(handler, SQSParser(encoding.JSONUnmarshal))
				_, err := v.handler.Invoke(context.Background(), []byte(tt.payload))
				assert.True(t, called)
				assert.NoError(t, err)
			})
		}
	})
}
