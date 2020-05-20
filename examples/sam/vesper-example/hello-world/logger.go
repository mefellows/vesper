package main

import (
	"context"
	"encoding/json"
	"log"
)

type ctxKey string

const (
	// CtxKeyCorrelationID is the context key for correlating IDs across logs
	CtxKeyCorrelationID = ctxKey("correlationId")
)

type logMessage struct {
	Message       interface{}
	Loglevel      string
	CorrelationID string
}

type customLogger struct {
	ctx      context.Context
	logLevel string
}

func newLogger() *customLogger {
	return &customLogger{
		logLevel: "DEBUG",
		ctx:      context.Background(),
	}
}

func (l *customLogger) WithContext(ctx context.Context) {
	l.ctx = ctx
}

func (l customLogger) Println(i ...interface{}) {

	value := l.ctx.Value(CtxKeyCorrelationID)
	correlationID, ok := value.(string)

	if !ok {
		correlationID = ""
	}

	m := logMessage{
		CorrelationID: correlationID,
		Loglevel:      l.logLevel,
		Message:       i,
	}

	s, _ := json.Marshal(m)

	log.Println(string(s))
}
