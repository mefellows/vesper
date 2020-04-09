# Vesper - Middleware for AWS Lambda

<!-- TOC -->

- [Vesper - Middleware for AWS Lambda](#vesper---middleware-for-aws-lambda)
  - [Get started](#get-started)
  - [API](#api)
    - [Usage](#usage)
    - [Logging](#logging)
  - [Middleware](#middleware)
    - [Warmup](#warmup)
  - [What's in a name?](#whats-in-a-name)

<!-- /TOC -->

## Get started

Create a new serverless project from the Vesper template:

```sh
serverless create -u https://github.com/mefellows/vesper/tree/master/template
```

## API

### Usage

```go
// MyHandler implements the Lambda Handler interface
func MyHandler(ctx context.Context, u User) (interface{}, error) {
	log.Println("[MyHandler]: handler invoked with user: ", u.Username)

	return u.Username, nil
}

func main() {
  // Create a new vesper instance, passing in all the Middlewares
  v := vesper.New(MyHandler, middleware.WarmupMiddleware)

  // Replace the standard lambda.Start() with Vespers wrapper
	v.Start()
}
```

### Logging

You can set your own custom logger with `vesper.Logger(l LogPrinter)`.

## Auto unmarshalling

The default behavior for Vesper is to automatically JSON unmarshal the payload into the type specificed in the handler parameter. This is consistent with the behaviour of the AWS Go Lambda library. This is useful if your handler accepts an input parameter which can be directly JSON unmarshalled into the parameter type. An example of this is the event types found in `github.com/aws/aws-lambda-go/events`.

This behaviour can be turned off in which case Vesper will start the middleware chain with the payload in `[]byte` form. If you are using a Parser middleware, you will likely want to turn this off to allow the parser to do the converting instead of Vesper.

Here is an example of how to turn off the auto unmarshalling:

```go
// MyHandler implements the Lambda Handler interface
func MyHandler(ctx context.Context, input []byte) error {
  log.Println("[MyHandler]: handler invoked with payload: ", input)

	return nil
}

func main() {
  // Create a new vesper instance and disable auto unmarshalling
  v := vesper.New(MyHandler).
  			DisableAutoUnmarshal()
  
  // Replace the standard lambda.Start() with Vespers wrapper
	v.Start()
}
```



## Middleware

### Warmup

Short circuits a request if the serverless warmup event is detected.

**TIP: This middleware should be included early in the chain, before any validation or processing happens**

Implements a warmup handler for https://www.npmjs.com/package/serverless-plugin-warmup

Example:

```go
func main() {
	m := vesper.New(MyHandler, vesper.WarmupMiddleware, /* any other middlewares here */)
	m.Start()
}
```

### Parser

Parses the input payload to the type specificed in the handler parameter. It accepts a decoder function so you can decide how it parses the payload.

**NOTE: the auto unmarshaling needs to be turned off for this middleware to work correctly. See [Auto unmarshalling](#auto-unmarshalling).**

Example of usage:

```go
import (
	"encoding/json"

	"github.com/mefellows/vesper"
)

func main() {
	m := vesper.New(MyHandler).
				DisableAutoUnmarshal().
  			Use(vesper.ParserMiddleware(json.Unmarshal))
	m.Start()
}
```

### JSONParser

This is a shorthand for using the [Parser middleware](#parser) - `vesper.ParserMiddleware(json.Unmarshal)`.

Example of usage:

```go
import "github.com/mefellows/vesper"

func main() {
	m := vesper.New(MyHandler).
				DisableAutoUnmarshal().
  			Use(vesper.JSONParserMiddleware())
	m.Start()
}
```

### SQSParser

Parses the input payload as an SQS event and then extracts and unmarshals the body into the type specificed in the handler parameter. The handler parameter must be a slice as SQS events are always batched. It accepts a decoder function so you can decide how it parses the SQS record body.

**NOTE: the auto unmarshaling needs to be turned off for this middleware to work correctly. See [Auto unmarshalling](#auto-unmarshalling).**

Example of usage:

```go
import (
	"encoding/json"

	"github.com/mefellows/vesper"
)

func MyHandler(ctx context.Context, users []users) error {
  log.Println("[MyHandler]: handler invoked with users: ", input)

	return nil
}

func main() {
	m := vesper.New(MyHandler).
				DisableAutoUnmarshal().
  			Use(vesper.SQSParserMiddleware(json.Unmarshal))
	m.Start()
}
```

### JSONSQSParser

This is a shorthand for using the [SQSParser middleware](#sqsparser) - `vesper.SQSParserMiddleware(json.Unmarshal)`.

Example of usage:

```go
import "github.com/mefellows/vesper"

func main() {
	m := vesper.New(MyHandler).
				DisableAutoUnmarshal().
  			Use(vesper.JSONParserMiddleware())
	m.Start()
}
```

## What's in a name?

Golang has a rich history of naming HTTP and middleware type libraries on classy gin-based beverages (think Gin, Martini and Negroni). Vesper is yet another gin-based beverage.

Vesper was also inspired by [Middy JS](https://github.com/middyjs/middy), who's mascot is a Moped.

Vesper is a (not so clever) _portmanteau_ of Vesper, the gin-based martini, and Vespa, a beautiful Italian scooter.
