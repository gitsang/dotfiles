package service

import (
	"context"
	"log/slog"
	"strings"
	"time"
)

type EmptyMessageError struct {
	Code    string
	Message string
}

var ErrEmptyMessage = EmptyMessageError{
	Code:    "EMPTY_MESSAGE",
	Message: "message must not be empty",
}

func (e EmptyMessageError) Error() string {
	return e.Message
}

func (e EmptyMessageError) Is(target error) bool {
	_, ok := target.(EmptyMessageError)
	return ok
}

// EchoService contains echo use cases independent from transport details.
type EchoService struct {
	logHandler slog.Handler
	repository EchoRepository
}

func NewEchoService(logHandler slog.Handler, repository EchoRepository) *EchoService {
	return &EchoService{logHandler: logHandler, repository: repository}
}

func (s *EchoService) Echo(ctx context.Context, message string) (string, error) {
	entryTime := time.Now()
	logger := slog.New(s.logHandler).With(slog.String("message", message))
	defer func() {
		logger.Debug("echo completed", slog.String("cost", time.Since(entryTime).String()))
	}()

	if strings.TrimSpace(message) == "" {
		return "", ErrEmptyMessage
	}
	if err := s.repository.Save(ctx, message); err != nil {
		return "", err
	}

	return message, nil
}
