package pubsubworker

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/pubsub/v2"
)

func (w *Worker) SubscribeUserCreated(ctx context.Context, subName string) func(context.Context) {
	return func(ctx context.Context) {
		w.log.Infof(ctx, "Listening user-created subscription: %s", subName)

		w.client.Subscribe(
			ctx,
			subName,
			func(ctx context.Context, msg *pubsub.Message) error {
				// 👉 logic bisnis di sini
				time.Sleep(200 * time.Millisecond)
				fmt.Printf("Received user created: %s\n", string(msg.Data))
				return nil
			},
		)
	}
}
