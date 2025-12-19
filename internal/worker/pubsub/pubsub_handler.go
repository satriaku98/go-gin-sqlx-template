package pubsubworker

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/pubsub/v2"
)

func (w *Worker) SubscribeUserCreated(ctx context.Context, subName string) func(context.Context) {
	return func(ctx context.Context) {
		defer func() {
			if r := recover(); r != nil {
				w.log.Errorf(ctx, "panic in pubsub handler: %v", r)
			}
		}()

		w.log.Infof(ctx, "Listening user-created subscription: %s", subName)

		w.client.Subscribe(
			ctx,
			subName,
			func(ctx context.Context, msg *pubsub.Message) error {
				// implement your business logic here
				time.Sleep(200 * time.Millisecond) // simulate processing time
				fmt.Printf("Received user created: %s\n", string(msg.Data))
				return nil
			},
		)
	}
}
