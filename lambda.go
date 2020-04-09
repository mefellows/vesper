package vesper

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
)

// This is the raw Lambda handler interface AWS needs to run a Lambda function
type lambdaHandler func(context.Context, []byte) (interface{}, error)

// Invoke calls the handler, and serializes the response.
// If the underlying handler returned an error, or an error occurs during serialization, error is returned.
func (handler lambdaHandler) Invoke(ctx context.Context, payload []byte) ([]byte, error) {
	response, err := handler(ctx, payload)
	if err != nil {
		return nil, err
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		return nil, err
	}

	return responseBytes, nil
}

func errorHandler(e error) lambdaHandler {
	return func(ctx context.Context, event []byte) (interface{}, error) {
		return nil, e
	}
}

func validateHandlerFunc(handlerInterface interface{}) error {
	handlerType, err := handlerType(handlerInterface)
	if err != nil {
		return err
	}
	if err = validateArguments(handlerType); err != nil {
		return err
	}
	if err = validateReturns(handlerType); err != nil {
		return err
	}
	return nil
}

func handlerType(handlerInterface interface{}) (reflect.Type, error) {
	if handlerInterface == nil {
		return nil, fmt.Errorf("handler is nil")
	}
	handlerType := reflect.TypeOf(handlerInterface)
	if handlerType.Kind() != reflect.Func {
		return nil, fmt.Errorf("handler kind %s is not %s", handlerType.Kind(), reflect.Func)
	}
	return handlerType, nil
}

func validateArguments(handler reflect.Type) error {
	if handler.NumIn() > 2 {
		return fmt.Errorf("handlers may not take more than two arguments, but handler takes %d", handler.NumIn())
	} else if handler.NumIn() > 0 {
		contextType := reflect.TypeOf((*context.Context)(nil)).Elem()
		argumentType := handler.In(0)
		handlerTakesContext := argumentType.Implements(contextType)
		if handler.NumIn() > 1 && !handlerTakesContext {
			return fmt.Errorf("handler takes two arguments, but the first is not Context. got %s", argumentType.Kind())
		}
	}

	return nil
}

func validateReturns(handler reflect.Type) error {
	errorType := reflect.TypeOf((*error)(nil)).Elem()
	if handler.NumOut() > 2 {
		return fmt.Errorf("handler may not return more than two values")
	} else if handler.NumOut() > 1 {
		if !handler.Out(1).Implements(errorType) {
			return fmt.Errorf("handler returns two values, but the second does not implement error")
		}
	} else if handler.NumOut() == 1 {
		if !handler.Out(0).Implements(errorType) {
			return fmt.Errorf("handler returns a single value, but it does not implement error")
		}
	}
	return nil
}

func handlerTakesContext(handlerType reflect.Type) bool {
	if handlerType.NumIn() > 0 {
		contextType := reflect.TypeOf((*context.Context)(nil)).Elem()
		argumentType := handlerType.In(0)
		return argumentType.Implements(contextType)
	}
	return false
}
