package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go-gin-sqlx-template/config"
	ps "go-gin-sqlx-template/internal/integration/pubsub"

	"cloud.google.com/go/pubsub/v2"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig(".")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if cfg.PubSubProjectID == "" {
		log.Fatal("PUBSUB_PROJECT_ID is not set in config")
	}

	options := []option.ClientOption{}
	if cfg.PubSubCredsFile != "" {
		log.Println("Using credentials file:", cfg.PubSubCredsFile)
		options = append(options, option.WithCredentialsFile(cfg.PubSubCredsFile))
	}
	if cfg.PubSubEmulatorHost != "" {
		log.Println("Using emulator host:", cfg.PubSubEmulatorHost)
		options = append(options,
			option.WithEndpoint(cfg.PubSubEmulatorHost),
			option.WithoutAuthentication(),
			option.WithGRPCDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
		)
	}

	ctx := context.Background()

	// Initialize Pub/Sub client
	client, err := ps.NewClient(ctx, cfg.PubSubProjectID, options...)
	if err != nil {
		log.Fatalf("Failed to create pubsub client: %v", err)
	}
	defer client.Close()

	topicID := "example-topic"
	subID := "example-sub"

	// Create Topic
	if err := client.EnsureTopic(ctx, topicID); err != nil {
		log.Fatalf("Failed to create topic: %v", err)
	}
	fmt.Printf("Topic %s is ready\n", topicID)

	// Create Subscription
	if err := client.EnsureSubscription(ctx, subID, topicID); err != nil {
		log.Fatalf("Failed to create subscription: %v", err)
	}
	fmt.Printf("Subscription %s is ready\n", subID)

	// Subscribe in a goroutine
	go func() {
		fmt.Println("Listening for messages...")
		client.Subscribe(
			ctx,
			subID,
			func(ctx context.Context, msg *pubsub.Message) error {
				// process
				fmt.Printf("Received message: %s\n", string(msg.Data))
				return nil // Ack
			},
			ps.WithReceiveSettings(10),
		)
	}()

	// Publish a message
	time.Sleep(2 * time.Second) // Wait for subscriber to be ready
	msg := "Hello, Pub/Sub!"
	fmt.Printf("Publishing message: %s\n", msg)
	id, err := client.Publish(ctx, topicID, []byte(msg), nil)
	if err != nil {
		log.Fatalf("Failed to publish message: %v", err)
	}
	fmt.Printf("Message published with ID: %s\n", id)

	// Wait a bit to ensure message is received
	time.Sleep(5 * time.Second)
}
