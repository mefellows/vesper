package spike

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/mefellows/vesper"
)

type Event struct {
	ID     string      `json:"event_id"`
	Source string      `json:"source"`
	Data   interface{} `json:"data"`
}

var eventDetails = func(f vesper.LambdaFunc) vesper.LambdaFunc {
	return func(ctx context.Context, in interface{}) (interface{}, error) {
		log.Println("[eventDetails] start")

		// update context...

		res, err := f(ctx, in)
		log.Println("[eventDetails] END: ", res)

		return res, err
	}
}

type Event struct {
	context.Context
}
type User struct {
	Event
	Username string
	Name     string
}

var kinesisParser = func(f vesper.LambdaFunc) vesper.LambdaFunc {
	return func(ctx context.Context, in interface{}) (interface{}, error) {
		log.Println("[kinesisParser] start")

		kinesisEvent := in.(events.KinesisEvent)
		// var kinesisParsed OurKinesisWrapperType

		for _, record := range kinesisEvent.Records {
			kinesisRecord := record.Kinesis
			dataBytes := kinesisRecord.Data
			dataText := string(dataBytes)

			fmt.Printf("%s Data = %s \n", record.EventName, dataText)
		}

		// update context...

		res, err := f(ctx, kinesisEvent)
		log.Println("[kinesisParser] END: ", res)

		return res, err
	}
}

type Event struct {
	context.Context
	// Josh proposal here
}

type UserEvent struct {
	Event
	User
}

// MyHandler implements the Lambda Handler interface
func MyHandler(ctx context.Context, users []UserEvent) (interface{}, error) {
	for _, user := range users {
		logger.WithContext(user.Context).Info("aoeuaoeu")

		// Do something
	}

	// log.Println("[actual handler]: Have User %+v\n", u)

	return nil, nil
	// return u.Username, nil
}

func main() {
	m := vesper.New(MyHandler, authentication, kinesisHandler)
	m.Start()
}

// # TODO
// 1. Implement HandlerSignatureMiddleware
// 3. Implement SQS Typed Handler Middleware
// 4. Cleanup / write tests for Vesper
// 5. Write / Publish documentation
// 6. Uplift Pkg logging to support lambda? (or new repo/library)
// 7. Setup CI for Vesper
// 8. Integrate / demo with lambda starter kit?

// 2. Implement Kinesis Typed Handler Middleware

// Principles

// 1. Preserve type safety
// 1. Stay as close as possible to AWS Go library to ensure compatibility with tools like SAM, Serverless, local testing and so on, as well as reduce cognitive overload
// 1. Allow user to controls message batch semantics (e.g. ability to control concurrency)
// 1. Avoid too much magic
// 1. Keep vesper as middleware library only
// 1. Enable/Allow the user of proposed Latitude standardised Message structure - https://di.latitudefinancial.com/wiki/pages/viewpage.action?spaceKey=EA&title=Event+Message+Structure+-+PROPOSAL
