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

func SQSParserMiddleware(unmarshaler encoding.UnmarshalFunc) func(LambdaFunc) LambdaFunc {

	validateTIn := func(tIn reflect.Type, ok bool) error {
		if !ok {
			return errors.New("no input parameter in lambda handler signature")
		}
		if tIn == reflect.TypeOf(events.SQSEvent{}) {
			return errors.New("SQSParserMiddleware middleware should not be used if input parameter is events.SQSEvent")
		}
		if tIn.Kind() != reflect.Slice {
			return errors.New("input parameter for SQS event must be a slice")
		}
		return nil
	}

	unmarshalSQSEventFromContext := func(ctx context.Context) (events.SQSEvent, error) {
		bs, ok := PayloadFromContext(ctx)
		if !ok {
			return events.SQSEvent{}, errors.New("payload not available in context. this is likely a bug with the framework")
		}
		evt := events.SQSEvent{}
		if err := json.Unmarshal(bs, &evt); err != nil {
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
			if err := validateTIn(tIn, ok); err != nil {
				return nil, err
			}
			evt := events.SQSEvent{}
			evt, err := unmarshalSQSEventFromContext(ctx)
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

func JSONSQSParserMiddleware() func(LambdaFunc) LambdaFunc {
	return SQSParserMiddleware(encoding.JSONUnmarshal)
}
