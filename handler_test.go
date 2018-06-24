package main

import (
	"context"
	"fmt"
	"testing"
)

func Test_buildTypedMiddleware(t *testing.T) {
	m := New(MyHandler, AuthMiddleware, DummyMiddleware)
	ctx := context.TODO()

	// The Lamda handler interface takes a context and a byte[]
	i := []byte(`{"username":"matt", "password":"nottelling"}`)
	res, err := m.handler(ctx, i)

	fmt.Printf("res: %v, err: %v", res, err)
}
