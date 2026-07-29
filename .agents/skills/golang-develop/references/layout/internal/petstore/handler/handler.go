package handler

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang-develop-example/internal/petstore/service"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

type Handler struct {
	service    *service.Collection
	logHandler slog.Handler
}

type Server struct {
	logHandler   slog.Handler
	httpListener net.Listener
	httpServer   *http.Server
	grpcListener net.Listener
	grpcServer   *grpc.Server
	service      *service.Collection
	apiPrefix    string
	httpTimeout  time.Duration
	stopOnce     sync.Once
	stopErr      error
}

type OptionFunc func(*Server)

func WithLogHandler(handler slog.Handler) OptionFunc {
	return func(s *Server) {
		s.logHandler = handler
	}
}

func WithHTTPListener(listener net.Listener) OptionFunc {
	return func(s *Server) {
		s.httpListener = listener
	}
}

func WithGRPCListener(listener net.Listener) OptionFunc {
	return func(s *Server) {
		s.grpcListener = listener
	}
}

func WithServiceCollection(collection *service.Collection) OptionFunc {
	return func(s *Server) {
		s.service = collection
	}
}

func WithAPIPrefix(prefix string) OptionFunc {
	return func(s *Server) {
		s.apiPrefix = prefix
	}
}

func WithHTTPTimeout(timeout time.Duration) OptionFunc {
	return func(s *Server) {
		s.httpTimeout = timeout
	}
}

func New(optfs ...OptionFunc) *Server {
	server := &Server{
		logHandler: slog.NewTextHandler(io.Discard, nil),
		apiPrefix:  "/api",
	}
	for _, apply := range optfs {
		if apply != nil {
			apply(server)
		}
	}
	return server
}

func (s *Server) initServer(ctx context.Context) error {
	handler := &Handler{
		service:    s.service,
		logHandler: s.logHandler,
	}

	s.grpcServer = grpc.NewServer()
	RegisterServiceServer(s.grpcServer, handler)

	apiRouter := runtime.NewServeMux()
	if err := RegisterServiceHandlerServer(ctx, apiRouter, handler); err != nil {
		return fmt.Errorf("register service handler: %w", err)
	}

	topRouter := http.NewServeMux()
	prefix := "/" + strings.Trim(s.apiPrefix, "/")
	if prefix == "/" {
		topRouter.Handle("/", apiRouter)
	} else {
		topRouter.Handle(prefix+"/", http.StripPrefix(prefix, apiRouter))
		topRouter.HandleFunc(prefix, func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, prefix+"/", http.StatusTemporaryRedirect)
		})
	}

	s.httpServer = &http.Server{
		Handler:           topRouter,
		ReadHeaderTimeout: s.httpTimeout,
	}

	return nil
}

func (s *Server) Serve(ctx context.Context) error {
	logger := slog.New(s.logHandler)
	if err := s.initServer(ctx); err != nil {
		return err
	}

	errCh := make(chan error, 2)

	logger.Info("starting gRPC server", slog.String("addr", s.grpcListener.Addr().String()))
	go func() {
		if err := s.grpcServer.Serve(s.grpcListener); err != nil {
			errCh <- fmt.Errorf("gRPC server error: %w", err)
		}
	}()

	logger.Info("starting HTTP server", slog.String("addr", s.httpListener.Addr().String()))
	go func() {
		if err := s.httpServer.Serve(s.httpListener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("HTTP server error: %w", err)
		}
	}()

	select {
	case err := <-errCh:
		_ = s.Stop(context.Background())
		return err
	case <-ctx.Done():
		return s.Stop(context.Background())
	}
}

func (s *Server) Stop(ctx context.Context) error {
	s.stopOnce.Do(func() {
		if s.grpcServer != nil {
			s.grpcServer.GracefulStop()
		}
		if s.httpServer != nil {
			s.stopErr = errors.Join(s.stopErr, s.httpServer.Shutdown(ctx))
		}
	})
	return s.stopErr
}
