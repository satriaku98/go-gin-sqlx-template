package pubsubworker

import (
	"context"

	"go-gin-sqlx-template/pkg/logger"
	ps "go-gin-sqlx-template/pkg/pubsub"
)

type Worker struct {
	client *ps.Client
	log    *logger.Logger
}

func New(client *ps.Client, log *logger.Logger) *Worker {
	return &Worker{
		client: client,
		log:    log,
	}
}

func (w *Worker) Start(ctx context.Context, subs ...func(context.Context)) {
	for _, sub := range subs {
		go sub(ctx)
	}
}
