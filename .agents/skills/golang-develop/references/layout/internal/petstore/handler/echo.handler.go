package handler

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"golang-develop-example/internal/petstore/service"
	api "golang-develop-example/pkg/api/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type EchoHandler struct {
	api.UnimplementedEchoServiceServer
	*Handler
}

func NewEchoHandler(handler *Handler) *EchoHandler {
	return &EchoHandler{Handler: handler}
}

func (h *EchoHandler) Echo(ctx context.Context, req *api.EchoRequest) (*api.EchoResponse, error) {
	entryTime := time.Now()
	logger := slog.New(h.logHandler).With(slog.Any("req", req))
	defer func() {
		logger.Debug("end", slog.String("cost", time.Since(entryTime).String()))
	}()

	message, err := h.service.Echo.Echo(ctx, req.GetMessage())
	if err != nil {
		logger.Error("service failed", slog.Any("err", err))
		switch {
		case errors.Is(err, service.ErrEmptyMessage):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		default:
			return nil, status.Error(codes.Unknown, err.Error())
		}
	}

	return &api.EchoResponse{Message: message}, nil
}
