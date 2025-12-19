package pubsub

import (
	"context"
	"fmt"
	"go-gin-sqlx-template/config"
	"sync"

	"cloud.google.com/go/pubsub/v2"
	"cloud.google.com/go/pubsub/v2/apiv1/pubsubpb"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
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
func NewClient(cfg config.Config) (*Client, error) {
	if cfg.PubSubProjectID == "" {
		return nil, fmt.Errorf("PUBSUB_PROJECT_ID is not set in config")
	}

	dialOpts := []grpc.DialOption{
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	}

	if cfg.PubSubEmulatorHost != "" {
		dialOpts = append(dialOpts,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
	}

	var options []option.ClientOption
	for _, opt := range dialOpts {
		options = append(options, option.WithGRPCDialOption(opt))
	}

	if cfg.PubSubCredsFile != "" {
		options = append(options, option.WithCredentialsFile(cfg.PubSubCredsFile))
	}
	if cfg.PubSubEmulatorHost != "" {
		options = append(options,
			option.WithEndpoint(cfg.PubSubEmulatorHost),
			option.WithoutAuthentication(),
		)
	}

	ctx := context.Background()

	c, err := pubsub.NewClient(ctx, cfg.PubSubProjectID, options...)
	if err != nil {
		return nil, fmt.Errorf("create pubsub client: %w", err)
	}

	return &Client{client: c}, nil
}

// EnsureAll ensures all topics and subscriptions exist.
// If any topic or subscription does not exist, it will be created.
// This method is intended to be called during application startup (fail-fast).
func (c *Client) EnsureAll(ctx context.Context, topics []TopicConfig) error {
	for _, t := range topics {
		if err := c.EnsureTopic(ctx, t.Topic); err != nil {
			return err
		}
		for _, sub := range t.Subs {
			if err := c.EnsureSubscription(ctx, sub, t.Topic); err != nil {
				return err
			}
		}
	}
	return nil
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

	// Inject trace context
	if attrs == nil {
		attrs = make(map[string]string)
	}
	otel.GetTextMapPropagator().Inject(ctx, propagation.MapCarrier(attrs))

	tr := otel.Tracer("pubsub")
	ctx, span := tr.Start(ctx, "publish "+topicID)
	span.SetAttributes(
		attribute.String("topic_id", topicID),
		attribute.String("data", string(data)),
		attribute.String("attributes", fmt.Sprintf("%v", attrs)),
	)
	defer span.End()

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
		if msg.Attributes != nil {
			carrier := propagation.MapCarrier(msg.Attributes)
			ctx = otel.GetTextMapPropagator().Extract(ctx, carrier)
		}

		tr := otel.Tracer("pubsub")
		ctx, span := tr.Start(ctx, "receive "+subscriptionID)
		span.SetAttributes(
			attribute.String("subscription_id", subscriptionID),
			attribute.String("data", string(msg.Data)),
			attribute.String("attributes", fmt.Sprintf("%v", msg.Attributes)),
		)
		defer span.End()

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
