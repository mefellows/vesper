TEST?=./...

default: test

deploy:
	GOOS=linux go build -ldflags="-s -w" -o bin/main
	sls deploy -f hello --region ap-southeast-2

deps:
	go get -d -v -p 2 ./...

.PHONY: bin default dev test updatedeps
