# UPDATE

Starting from [version 1.26](https://github.com/serverless/serverless/releases/tag/v1.26.0) Serverless Framework includes two Golang templates:

- `aws-go` - basic template with two functions
- `aws-go-dep` - **recommended** template using [`dep`](https://github.com/golang/dep) package manager

You can use them with `create` command:

```
serverless create -t aws-go-dep
```

Original README below.

---

# Serverless Template for Golang

This repository contains template for creating serverless services written in Golang.

## Quick Start

1.  Create a new service based on this template

```
serverless create -u https://github.com/serverless/serverless-golang/ -p myservice
```

2.  Compile function

```
cd myservice
GOOS=linux go build -o bin/main
```

3. Test Locally

```
sls invoke local -f hello -d '{ "Event": { "source": "serverless-plugin-warmup" } }'
sls invoke local -f hello --data '{"Username": "me", "Password":"securething"}'
```

4.  Deploy!

```
serverless deploy
```

## What's in a name?

Golang has a rich history of naming HTTP and middleware type libraries on classy gin-based beverages (think Gin, Martini and Negroni). Vesper is yet another gin-based beverage.

Vesper was also inspired by [Middy JS](https://github.com/middyjs/middy), who's mascot is a Moped.

Vesper is a (not so clever) _portmanteau_ of Vesper, the gin-based martini, and Vespa, a beautiful Italian scooter.
