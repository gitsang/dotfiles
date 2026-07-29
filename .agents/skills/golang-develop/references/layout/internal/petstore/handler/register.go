package handler

import (
	"context"

	api "golang-develop-example/pkg/api/v1"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

func RegisterServiceServer(grpcServer *grpc.Server, handler *Handler) {
	api.RegisterEchoServiceServer(grpcServer, NewEchoHandler(handler))
}

func RegisterServiceHandlerServer(ctx context.Context, apiRouter *runtime.ServeMux, handler *Handler) error {
	return api.RegisterEchoServiceHandlerServer(ctx, apiRouter, NewEchoHandler(handler))
}
