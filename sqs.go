package vesper

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/aws/aws-lambda-go/events"
	"github.com/mefellows/vesper/encoding"
)

// SQSParserMiddleware transforms SQS event records into the handler input parameter type using the given unmarshaler.
// The handler input parameter must be a slice.
func SQSParserMiddleware(unmarshaler encoding.UnmarshalFunc) func(LambdaFunc) LambdaFunc {
	validateTIn := func(tIn reflect.Type) error {
		if tIn == reflect.TypeOf(events.SQSEvent{}) {
			return errors.New("SQSParserMiddleware middleware should not be used if input parameter is events.SQSEvent")
		}
		if tIn.Kind() != reflect.Slice {
			return errors.New("input parameter for SQS event must be a slice")
		}
		return nil
	}

	unmarshalSQSEvent := func(in interface{}) (events.SQSEvent, error) {
		b, ok := in.([]byte)
		if !ok {
			return events.SQSEvent{}, fmt.Errorf("expected []byte input but got %T", in)
		}
		evt := events.SQSEvent{}
		if err := json.Unmarshal(b, &evt); err != nil {
			return events.SQSEvent{}, fmt.Errorf("could not unmarshal SQS event: %w", err)
		}
		return evt, nil
	}

	unmarshalRecords := func(tIn reflect.Type, evt events.SQSEvent) (reflect.Value, error) {
		tIns := reflect.MakeSlice(tIn, 0, len(evt.Records))
		for _, r := range evt.Records {
			msgBody, err := unmarshalToType(unmarshaler, tIn.Elem(), []byte(r.Body))
			if err != nil {
				return reflect.Value{}, fmt.Errorf("could not unmarshal SQS message body for ID %s: %w", r.MessageId, err)
			}
			tIns = reflect.Append(tIns, reflect.ValueOf(msgBody))
		}
		return tIns, nil
	}

	return func(next LambdaFunc) LambdaFunc {
		return func(ctx context.Context, in interface{}) (interface{}, error) {
			if unmarshaler == nil {
				return nil, errors.New("no unmarshaler was provided")
			}
			tIn, ok := TInFromContext(ctx)
			if !ok {
				return next(ctx, in) // continue as there is no TIn to parse anyway.
			}
			if err := validateTIn(tIn); err != nil {
				return nil, err
			}
			evt, err := unmarshalSQSEvent(in)
			if err != nil {
				return nil, err
			}
			tIns, err := unmarshalRecords(tIn, evt)
			if err != nil {
				return nil, err
			}
			return next(ctx, tIns.Interface())
		}
	}
}

// JSONSQSParserMiddleware transforms SQS event records into the handler input parameter type using a JSON unmarshaler.
// The handler input parameter must be a slice.
func JSONSQSParserMiddleware() func(LambdaFunc) LambdaFunc {
	return SQSParserMiddleware(json.Unmarshal)
}
