package repository

import (
	"context"
	"sync"
)

// EchoRepository stores echo messages behind the service contract.
type EchoRepository struct {
	mu       sync.Mutex
	messages []string
}

func NewEchoRepository() *EchoRepository {
	return &EchoRepository{}
}

func (r *EchoRepository) Save(ctx context.Context, message string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.messages = append(r.messages, message)
	return nil
}
