# Vesper - Middleware for AWS Lambda

<!-- TOC -->

- [Vesper - Middleware for AWS Lambda](#vesper---middleware-for-aws-lambda)
  - [Get started](#get-started)
  - [API](#api)
    - [Usage](#usage)
    - [Logging](#logging)
  - [Middleware](#middleware)
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

You can set your own custom logger with `WithLogger(l *log.Logger)`.

## Middleware

## What's in a name?

Golang has a rich history of naming HTTP and middleware type libraries on classy gin-based beverages (think Gin, Martini and Negroni). Vesper is yet another gin-based beverage.

Vesper was also inspired by [Middy JS](https://github.com/middyjs/middy), who's mascot is a Moped.

Vesper is a (not so clever) _portmanteau_ of Vesper, the gin-based martini, and Vespa, a beautiful Italian scooter.
