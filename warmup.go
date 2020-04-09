package vesper

import (
	"context"
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
func WarmupMiddleware(f LambdaFunc) LambdaFunc {
	return func(ctx context.Context, in interface{}) (interface{}, error) {
		log.Println("[warmupMiddleware] START")
		var event warmupEvent
		if err := ExtractType(ctx, &event); err == nil {
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
