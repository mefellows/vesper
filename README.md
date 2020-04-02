# Vesper - Middleware for AWS Lambda

<!-- TOC -->

- [Vesper - Middleware for AWS Lambda](#vesper---middleware-for-aws-lambda)
  - [Principles](#principles)
  - [Get started](#get-started)
  - [API](#api)
    - [Usage](#usage)
    - [Logging](#logging)
  - [Middleware](#middleware)
    - [Warmup](#warmup)
  - [TODO](#todo)
  - [What's in a name?](#whats-in-a-name)

<!-- /TOC -->

## Principles

1. Keep Vesper as middleware library only
1. Compatibility with the AWS Go SDK interface to ensure seamless integration with tools like SAM, Serverless, 1ocal testing and so on, as well as reduce cognitive overload forusers
1. Preserve type safety and encourage the use of types
1. Allow user to controls message batch semantics (e.g. ability to control concurrency)
1. Be comprehensible / avoid magic
1. Enable/Allow use of user-defined messages structures


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
## TODO

TODO
* [] Cleanup interface / write tests for Vesper
* [] Setup CI
* [] Implement HandlerSignatureMiddleware
* [] Implement Typed Record Handler Middleware for SQS
* [] Implement Typed Record Handler Middleware for Kinesis
* [] Implement Typed Record Handler Middleware for SNS
* [] Write / Publish documentation
* [] Integrate / demo with lambda starter kit (using Message structure proposal)

## What's in a name?

Golang has a rich history of naming HTTP and middleware type libraries on classy gin-based beverages (think Gin, Martini and Negroni). Vesper is yet another gin-based beverage.

Vesper was also inspired by [Middy JS](https://github.com/middyjs/middy), who's mascot is a Moped.

Vesper is a (not so clever) _portmanteau_ of Vesper, the gin-based martini, and Vespa, a beautiful Italian scooter.
