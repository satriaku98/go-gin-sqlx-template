package pubsub

import "cloud.google.com/go/pubsub/v2"

type SubscribeOption func(*pubsub.Subscriber)

func WithReceiveSettings(maxMsgs int) SubscribeOption {
	return func(s *pubsub.Subscriber) {
		s.ReceiveSettings.MaxOutstandingMessages = maxMsgs
	}
}
