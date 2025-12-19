package pubsub

import "go-gin-sqlx-template/config"

type TopicConfig struct {
	Topic string
	Subs  []string
}

func GetTopicConfig(cfg config.Config) []TopicConfig {
	return []TopicConfig{
		{
			Topic: cfg.PubSubTopicUserCreated,
			Subs: []string{
				cfg.PubSubSubscriptionUserCreated,
			},
		},
	}
}
