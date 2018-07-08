package middleware

import (
	"context"
	"encoding/json"
	"log"

	"github.com/mefellows/vesper"
)

// WarmupEvent is the manual event
// See https://www.npmjs.com/package/serverless-plugin-warmup
// {
//   "Event": {
//     "source": "serverless-plugin-warmup"
//   }
// }
type warmupEvent struct {
	Event struct {
		Source string
	}
}

// WarmupMiddleware detects a warmup invocation event from the
// plugin "serverless-plugin-warmup", and returns early if found
//
// See https://www.npmjs.com/package/serverless-plugin-warmup for more
var WarmupMiddleware = func(f vesper.LambdaFunc) vesper.LambdaFunc {
	return func(ctx context.Context, in interface{}) (interface{}, error) {
		log.Println("[warmupMiddleware] start")
		if v := ctx.Value(vesper.PAYLOAD{}); v != nil {
			var event warmupEvent
			json.Unmarshal(v.([]byte), &event)
			if event.Event.Source == "serverless-plugin-warmup" {
				log.Println("[warmupMiddleware] warmup event detected, exiting")
				return "warmup", nil
			}
		}

		res, err := f(ctx, in)
		log.Println("[warmupMiddleware] END: ", res)

		return res, err
	}
}
