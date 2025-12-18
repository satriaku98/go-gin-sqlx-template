package pubsub

import (
	"context"
	"fmt"
	"sync"

	"cloud.google.com/go/pubsub/v2"
	"cloud.google.com/go/pubsub/v2/apiv1/pubsubpb"
	"google.golang.org/api/option"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Client is a high-level wrapper around Google Cloud Pub/Sub v2 client.
// It manages publishers lifecycle and provides simplified publish/subscribe APIs.
type Client struct {
	client     *pubsub.Client
	publishers sync.Map // topicID -> *pubsub.Publisher
}

// NewClient creates a new Google Cloud Pub/Sub v2 client for the given project.
// Caller is responsible for calling Close() to release resources.
func NewClient(ctx context.Context, projectID string, opts ...option.ClientOption) (*Client, error) {
	c, err := pubsub.NewClient(ctx, projectID, opts...)
	if err != nil {
		return nil, fmt.Errorf("create pubsub client: %w", err)
	}

	return &Client{client: c}, nil
}

// EnsureTopic ensures the given topic exists.
// If the topic does not exist, it will be created.
// This method is intended to be called during application startup (fail-fast).
func (c *Client) EnsureTopic(ctx context.Context, topicID string) error {
	name := fmt.Sprintf("projects/%s/topics/%s", c.client.Project(), topicID)

	_, err := c.client.TopicAdminClient.GetTopic(
		ctx, &pubsubpb.GetTopicRequest{Topic: name},
	)
	if err == nil {
		return nil
	}
	if status.Code(err) != codes.NotFound {
		return fmt.Errorf("check topic: %w", err)
	}

	_, err = c.client.TopicAdminClient.CreateTopic(
		ctx, &pubsubpb.Topic{Name: name},
	)
	if err != nil {
		return fmt.Errorf("create topic: %w", err)
	}
	return nil
}

// EnsureSubscription ensures the given subscription exists for the specified topic.
// If the subscription does not exist, it will be created.
// This method assumes the topic already exists.
func (c *Client) EnsureSubscription(ctx context.Context, subID, topicID string) error {
	subName := fmt.Sprintf("projects/%s/subscriptions/%s", c.client.Project(), subID)
	topicName := fmt.Sprintf("projects/%s/topics/%s", c.client.Project(), topicID)

	_, err := c.client.SubscriptionAdminClient.GetSubscription(
		ctx, &pubsubpb.GetSubscriptionRequest{Subscription: subName},
	)
	if err == nil {
		return nil
	}
	if status.Code(err) != codes.NotFound {
		return fmt.Errorf("check subscription: %w", err)
	}

	_, err = c.client.SubscriptionAdminClient.CreateSubscription(
		ctx,
		&pubsubpb.Subscription{
			Name:  subName,
			Topic: topicName,
		},
	)
	if err != nil {
		return fmt.Errorf("create subscription: %w", err)
	}
	return nil
}

// publisher returns a cached publisher for the given topic.
// Publishers are lazily created and reused to support high-throughput publishing.
func (c *Client) publisher(topicID string) *pubsub.Publisher {
	if p, ok := c.publishers.Load(topicID); ok {
		return p.(*pubsub.Publisher)
	}

	p := c.client.Publisher(topicID)
	actual, _ := c.publishers.LoadOrStore(topicID, p)
	return actual.(*pubsub.Publisher)
}

// Publish publishes a message to the given topic and returns the server-assigned message ID.
// The topic is assumed to already exist (no admin RPC is performed here).
func (c *Client) Publish(
	ctx context.Context,
	topicID string,
	data []byte,
	attrs map[string]string,
) (string, error) {

	p := c.publisher(topicID)

	result := p.Publish(ctx, &pubsub.Message{
		Data:       data,
		Attributes: attrs,
	})

	id, err := result.Get(ctx)
	if err != nil {
		return "", fmt.Errorf("publish message: %w", err)
	}
	return id, nil
}

// Subscribe starts receiving messages from the given subscription.
//
// This method blocks until the provided context is canceled or a fatal error occurs.
//
// The handler function controls message acknowledgment:
//   - return nil   → message will be Acked
//   - return error → message will be Nacked
func (c *Client) Subscribe(
	ctx context.Context,
	subscriptionID string,
	handler func(context.Context, *pubsub.Message) error,
	opts ...SubscribeOption,
) error {

	s := c.client.Subscriber(subscriptionID)
	for _, opt := range opts {
		opt(s)
	}

	err := s.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		if err := handler(ctx, msg); err != nil {
			msg.Nack()
			return
		}
		msg.Ack()
	})

	if err != nil {
		return fmt.Errorf("receive messages: %w", err)
	}
	return nil
}

// Close closes the client.
func (c *Client) Close() error {
	c.publishers.Range(func(_, v any) bool {
		v.(*pubsub.Publisher).Stop()
		return true
	})
	return c.client.Close()
}
