package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/mefellows/vesper"
	"github.com/qri-io/jsonschema"
)

var namedMiddleware = func(name string) vesper.Middleware {
	return func(f vesper.LambdaFunc) vesper.LambdaFunc {
		return func(ctx context.Context, in interface{}) (interface{}, error) {
			logger.Println(fmt.Sprintf("START %s", name))

			res, err := f(ctx, in)

			logger.Println(fmt.Sprintf("END %s", name))

			return res, err
		}
	}
}

var correlationIdMiddleware = func(next vesper.LambdaFunc) vesper.LambdaFunc {
	return func(ctx context.Context, in interface{}) (interface{}, error) {
		logger.Println("START correlationIdMiddleware")

		// Introspects event, and extracts correlation IDs for common event types
		// Mutates context
		extractCorrelationIds := func(in interface{}) (context.Context, error) {
			evt := in.(events.APIGatewayProxyRequest)

			// Extract from Header, or fall back to request ID
			correlationID, ok := evt.Headers["X-Correlation-Id"]
			if !ok {
				correlationID = evt.RequestContext.RequestID
			}

			return context.WithValue(ctx, CtxKeyCorrelationID, correlationID), nil
		}

		ctx, err := extractCorrelationIds(in)
		if err != nil {
			logger.Println("unable to extract a correlation ID", err)
		}

		// Add correlation ID to context
		logger.WithContext(ctx)

		res, err := next(ctx, in)

		logger.Println("END correlationIdMiddleware")

		return res, err
	}
}

var validationMiddleware = func(next vesper.LambdaFunc) vesper.LambdaFunc {
	return func(ctx context.Context, in interface{}) (interface{}, error) {
		logger.Println("START validationMiddleware")
		evt := in.(events.APIGatewayProxyRequest)

		if evt.HTTPMethod != "GET" {
			validateBody := func(in interface{}) error {
				var schemaData = []byte(`{
				"title": "Person",
				"type": "object",
				"properties": {
						"firstName": {
								"type": "string"
						},
						"lastName": {
								"type": "string"
						}
				},
				"required": ["firstName", "lastName"]
			}`)

				rs := &jsonschema.RootSchema{}
				if err := json.Unmarshal(schemaData, rs); err != nil {
					return err
				}

				if errors, _ := rs.ValidateBytes([]byte(evt.Body)); len(errors) > 0 {
					return fmt.Errorf("Unable to validate payload: %+v", errors)
				}

				return nil
			}

			err := validateBody(in)

			if err != nil {
				return events.APIGatewayProxyResponse{
					Body:       fmt.Sprintf("Invalid request: %+v", err),
					StatusCode: 400,
				}, nil
			}
		}

		res, err := next(ctx, in)

		logger.Println("END validationMiddleware")

		return res, err
	}
}
