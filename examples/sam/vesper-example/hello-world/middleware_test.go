package main

import (
	"context"
	"encoding/base64"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
)

func TestValidationMiddleware(t *testing.T) {
	t.Run("400 Response", func(t *testing.T) {
		nextFunc := func(ctx context.Context, in interface{}) (interface{}, error) {
			t.Errorf("unexpected call to next func")
			return nil, nil
		}
		middleware := validationMiddleware(nextFunc)
		res, err := middleware(context.Background(), []byte("{}"))
		assert.Equal(t, 400, res.(events.APIGatewayProxyResponse).StatusCode)
		assert.NoError(t, err)
	})

	t.Run("200 Response", func(t *testing.T) {
		called := false
		nextFunc := func(ctx context.Context, in interface{}) (interface{}, error) {
			called = true
			return nil, nil
		}
		middleware := validationMiddleware(nextFunc)
		_, err := middleware(context.Background(), generateGatewayEvent(`'{"firstName":"matt", "lastName":"fellows"}`))

		assert.NoError(t, err)
		assert.True(t, called)
	})
}

func TestCorrelationIDMiddleware(t *testing.T) {
	t.Run("ID from API Gateway Request ID", func(t *testing.T) {
		id := ""
		nextFunc := func(ctx context.Context, in interface{}) (interface{}, error) {
			id = ctx.Value(CtxKeyCorrelationID).(string)
			return nil, nil
		}
		middleware := correlationIDMiddleware(nextFunc)
		evt := generateGatewayEvent("")
		delete(evt.Headers, "X-Correlation-Id")
		_, err := middleware(context.Background(), evt)

		assert.NoError(t, err)
		assert.Equal(t, "aws-request-id", id)
	})

	t.Run("ID from request header", func(t *testing.T) {
		id := ""
		nextFunc := func(ctx context.Context, in interface{}) (interface{}, error) {
			id = ctx.Value(CtxKeyCorrelationID).(string)
			return nil, nil
		}
		middleware := correlationIDMiddleware(nextFunc)
		_, err := middleware(context.Background(), generateGatewayEvent(""))

		assert.NoError(t, err)
		assert.Equal(t, "1234", id)
	})
}

func generateGatewayEvent(body string) events.APIGatewayProxyRequest {
	return events.APIGatewayProxyRequest{
		HTTPMethod: "POST",
		Body:       base64.StdEncoding.EncodeToString([]byte(body)),
		Headers: map[string]string{
			"X-Correlation-Id": "1234",
		},
		RequestContext: events.APIGatewayProxyRequestContext{
			RequestID: "aws-request-id",
		},
	}
}
