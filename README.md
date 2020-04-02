# Vesper - Middleware for AWS Lambda

The Golang Middleware engine for AWS Lambda.

[![Build Status](https://travis-ci.org/mefellows/vesper.svg?branch=master)](https://travis-ci.org/mefellows/vesper)
[![Coverage Status](https://coveralls.io/repos/github/mefellows/vesper/badge.svg?branch=HEAD)](https://coveralls.io/github/mefellows/vesper?branch=HEAD)
[![Go Report Card](https://goreportcard.com/badge/github.com/mefellows/vesper)](https://goreportcard.com/report/github.com/mefellows/vesper)
[![GoDoc](https://godoc.org/github.com/mefellows/vesper?status.svg)](https://godoc.org/github.com/mefellows/vesper)

<!-- TOC -->

- [Vesper - Middleware for AWS Lambda](#vesper---middleware-for-aws-lambda)
  - [Introduction](#introduction)
  - [Get started](#get-started)
  - [API](#api)
    - [Usage](#usage)
    - [Logging](#logging)
  - [Writing your own Middleware](#writing-your-own-middleware)
  - [Available Middleware](#available-middleware)
    - [Warmup](#warmup)
  - [How it works](#how-it-works)
    - [Execution order](#execution-order)
    - [Interrupt middleware execution early](#interrupt-middleware-execution-early)
  - [TODO](#todo)
  - [Developer Documentation](#developer-documentation)
    - [Goals](#goals)
    - [Contributing](#contributing)
  - [What's in a name?](#whats-in-a-name)

<!-- /TOC -->

## Introduction

Vesper is a very simple middleware engine for Lamda functions. If you are used to HTTP Web frameworks like Gorilla Mux and Go Kit, then you will be familiar with the concepts adopted in Vesper.

Middleware allows developers to isolate common technical concerns - such as input/output validation, logging and error handling - into functions that *decorate* the main business logic. This enables you to reuse these focus on writing code that remains clean, readable and easy to test and maintain.

<p align="center">
  <a href="https://asciinema.org/a/180671">
    <img width="880" src="https://raw.githubusercontent.com/mefellows/vesper/master/docs/demo.svg?sanitize=true">
  </a>
</p>

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
  v := vesper.New(MyHandler, vesper.WarmupMiddleware)

  // Replace the standard lambda.Start() with Vesper's wrapper
  v.Start()
}
```

### Logging

You can set your own custom logger with `vesper.Logger(l LogPrinter)`.

## Writing your own Middleware

A middleware is a function that takes a `LambdaFunc` and returns another `LambdaFunc`. A
`LambdaFunc` is simple a named type for the AWS Handler [signature](https://github.com/aws/aws-lambda-go/blob/master/lambda/entry.go#L37-L49).

Most middleware's do three things:

1. Modify or perform some action on the incoming request (such as validating the request)
2. Call the next middleware in the chain
3. Modify or perform some action on the outgoing response (such as validating the response)

Example:

```go
var dummyMiddleware = func(next vesper.LambdaFunc) vesper.LambdaFunc {
	// one time scope setup area for middleware - e.g. in-memory FIFO cache

	return func(ctx context.Context, in interface{}) (interface{}, error) {
    log.Println("[dummyMiddleware] START:")

    // (1) Modify the incoming request, or update the context before passing to the
    // next middleware in the chain

    // (2) You must call the next middleware in the chain if the request should proceed
    // and you want other middleware to execute
    res, err := next(ctx, in)

    // (3) Your last chance to modify the response before it is passed to any remaining
    // middleware in the chain
		log.Println("[dummyMiddleware] END:", in)

		return res, err
	}
}
```

## Available Middleware

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

## How it works

Vesper implements the classic *onion-like* middleware pattern, with some peculiar details.

![Vesper middleware engine diagram](/docs/middleware-schematic.png)

When you attach a new middleware this will wrap the business logic contained in the handler
in two separate steps.

When another middleware is attached this will wrap the handler again and it will be wrapped by
all the previously added middlewares in order, creating multiple layers for interacting with
the *request* (event) and the *response*.

This way the *request-response cycle* flows through all the middlewares, the
handler and all the middlewares again, giving the opportunity within every step to
modify or enrich the current request, context or the response.


### Execution order

Middlewares have two phases: `before` and `after`.

The `before` phase, happens *before* the handler is executed. In this code the
response is not created yet, so you will have access only to the request.

The `after` phase, happens *after* the handler is executed. In this code you will
have access to both the request and the response.

If you have three middlewares attached as in the image above this is the expected
order of execution:

 - `middleware1` (before)
 - `middleware2` (before)
 - `middleware3` (before)
 - `handler`
 - `middleware3` (after)
 - `middleware2` (after)
 - `middleware1` (after)

Notice that in the `after` phase, middlewares are executed in reverse order,
this way the first handler attached is the one with the highest priority as it will
be the first able to change the request and last able to modify the response before
it gets sent to the user.

### Interrupt middleware execution early

Some middlewares might need to stop the whole execution flow and return a response immediately.

If you want to do this you cansimple omit invoking `next` middleware and return early.

*Note*: this will stop the execution of successive middlewares in any phase (before and after) and returns an early response (or an error) directly at the Lambda level. If your middlewares does a specific task on every request like output serialization or error handling, these won't be invoked in this case.

In this example we can use this capability for rejecting an unauthorised request:

```go
var authMiddleware = func(next vesper.LambdaFunc) vesper.LambdaFunc {
	return func(ctx context.Context, in interface{}) (interface{}, error) {
		log.Println("[authMiddleware] START: ", in)
		user := in.(User)
		if user.Username == "fail" {
			error := map[string]string{
				"error": "unauthorised",
      }

      // NOTE: we do not call the "next" middleware, and completely prevent execution of subsequent middlewares
			return error, fmt.Errorf("user %v is unauthorised", in)
		}

		res, err := next(ctx, in)
		log.Printf("[authMiddleware] END: %+v \n", in)

		return res, err
	}
}
```

## TODO

- [ ] Cleanup interface / write tests for Vesper
- [ ] Setup CI
- [ ] Implement HandlerSignatureMiddleware
- [ ] Implement Typed Record Handler Middleware for SQS
- [ ] Implement Typed Record Handler Middleware for Kinesis
- [ ] Implement Typed Record Handler Middleware for SNS
- [ ] Write / Publish documentation
- [ ] Integrate / demo with lambda starter kit (using Message structure proposal)

## Developer Documentation

### Goals

1. Vesper is a middleware library - it shall provide a small API for this purpose, along with common middlewares
1. Compatibility with the AWS Go SDK interface to ensure seamless integration with tools like SAM, Serverless, 1ocal testing and so on, and to reduce cognitive overload for users
1. Preserve type safety and encourage the use of types throughout the system
1. Allow user to controls message batch semantics (e.g. ability to control concurrency)
1. Be comprehensible / avoid magic
1. Enable/allow use of user-defined messages structures

### Contributing

See [CONTRIBUTING](CONTRIBUTING.md).

## What's in a name?

Golang has a rich history of naming HTTP and middleware type libraries on classy gin-based beverages (think Gin, Martini and Negroni). Vesper is yet another gin-based beverage.

Vesper was also inspired by [Middy JS](https://github.com/middyjs/middy), who's mascot is a Moped.

Vesper is a (not so clever) _portmanteau_ of Vesper, the gin-based martini, and Vespa, a beautiful Italian scooter.
