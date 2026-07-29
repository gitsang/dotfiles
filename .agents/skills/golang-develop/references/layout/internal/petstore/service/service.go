package service

import (
	"context"
	"io"
	"log/slog"
)

type EchoRepository interface {
	Save(ctx context.Context, message string) error
}

// Collection holds all application use cases owned by the petstore application.
type Collection struct {
	logHandler     slog.Handler
	echoRepository EchoRepository

	Echo *EchoService
}

type CollectionOptionFunc func(*Collection)

func WithLogHandler(handler slog.Handler) CollectionOptionFunc {
	return func(c *Collection) {
		c.logHandler = handler
	}
}

func WithEchoRepository(repository EchoRepository) CollectionOptionFunc {
	return func(c *Collection) {
		c.echoRepository = repository
	}
}

func NewCollection(optfs ...CollectionOptionFunc) *Collection {
	collection := &Collection{
		logHandler: slog.NewTextHandler(io.Discard, nil),
	}
	for _, apply := range optfs {
		if apply != nil {
			apply(collection)
		}
	}

	collection.Echo = NewEchoService(collection.logHandler, collection.echoRepository)
	return collection
}
