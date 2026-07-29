package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"strings"
	"sync"
	"time"

	"golang-develop-example/internal/petstore/config"
	"golang-develop-example/internal/petstore/handler"
	"golang-develop-example/internal/petstore/repository"
	"golang-develop-example/internal/petstore/service"
	"golang-develop-example/pkg/util/stringx"

	ycache "gitcode.yealink.com/server/devops/yops/ykit/cache"
	"gitcode.yealink.com/server/devops/yops/ykit/cache/pkg/memoryc"
	"gitcode.yealink.com/server/devops/yops/ykit/cache/pkg/redisc"
	"gitcode.yealink.com/server/devops/yops/ykit/mq"
	"gitcode.yealink.com/server/devops/yops/ykit/mq/memq"
	"gitcode.yealink.com/server/devops/yops/ykit/mq/redismq"
	"github.com/go-redis/redis/v8"
)

type App struct {
	ctx               context.Context
	cache             ycache.Cache
	messageQueue      mq.MessageQueue
	messageQueueClose func() error
	cancel            context.CancelFunc
	server            *handler.Server
	logger            *slog.Logger
	shutdownOnce      sync.Once
	shutdownErr       error
}

func New(ctx context.Context, cfg *config.Config, logger *slog.Logger) (*App, error) {
	ctx, cancel := context.WithCancel(ctx)

	var cacheClient ycache.Cache
	switch strings.ToLower(strings.TrimSpace(cfg.Data.Cache.Driver)) {
	case "memory":
		c, err := memoryc.New(memoryc.WithNamespaces(cfg.Data.Cache.Memory.Namespaces...))
		if err != nil {
			cancel()
			return nil, err
		}
		cacheClient = c
	case "redis":
		client := redis.NewUniversalClient(&redis.UniversalOptions{
			Addrs:      stringx.TrimSplit(cfg.Data.Cache.Redis.Addresses, ","),
			Username:   cfg.Data.Cache.Redis.Username,
			Password:   cfg.Data.Cache.Redis.Password,
			DB:         cfg.Data.Cache.Redis.Db,
			MasterName: cfg.Data.Cache.Redis.MasterName,
		})
		c, err := redisc.New(
			redisc.WithClient(client),
			redisc.WithNamespaces(cfg.Data.Cache.Redis.Namespaces...),
		)
		if err != nil {
			cancel()
			return nil, errors.Join(client.Close(), err)
		}
		cacheClient = c
	default:
		cancel()
		return nil, fmt.Errorf("unsupported cache driver %q", cfg.Data.Cache.Driver)
	}

	var messageQueue mq.MessageQueue
	var messageQueueClose func() error
	switch strings.ToLower(strings.TrimSpace(cfg.Data.Mq.Driver)) {
	case "memory":
		logh := slog.NewTextHandler(io.Discard, nil)
		opts := []memq.OptionFunc{
			memq.WithContext(ctx),
			memq.WithLogHandler(logh),
			memq.WithNamespaces(cfg.Data.Mq.Memory.Namespaces),
		}

		if value := strings.TrimSpace(cfg.Data.Mq.KeepaliveInterval); value != "" {
			keepalive, err := time.ParseDuration(value)
			if err != nil {
				cancel()
				return nil, errors.Join(cacheClient.Close(), fmt.Errorf("parse data.mq.keepaliveInterval: %w", err))
			}
			opts = append(opts, memq.WithKeepInterval(keepalive))
		}

		if value := strings.TrimSpace(cfg.Data.Mq.ConsumeTimeout); value != "" {
			consumeTimeout, err := time.ParseDuration(value)
			if err != nil {
				cancel()
				return nil, errors.Join(cacheClient.Close(), fmt.Errorf("parse data.mq.consumeTimeout: %w", err))
			}
			opts = append(opts, memq.WithConsumeTimeout(consumeTimeout))
		}

		if cfg.Data.Mq.MaxRetry > 0 {
			opts = append(opts, memq.WithMaxRetry(cfg.Data.Mq.MaxRetry))
		}
		messageQueue = memq.NewMemMq(opts...)
	case "redis":
		client := redis.NewUniversalClient(&redis.UniversalOptions{
			Addrs:      stringx.TrimSplit(cfg.Data.Mq.Redis.Addresses, ","),
			Username:   cfg.Data.Mq.Redis.Username,
			Password:   cfg.Data.Mq.Redis.Password,
			DB:         cfg.Data.Mq.Redis.Db,
			MasterName: cfg.Data.Mq.Redis.MasterName,
		})
		if _, err := client.Ping(ctx).Result(); err != nil {
			cancel()
			return nil, errors.Join(cacheClient.Close(), client.Close(), fmt.Errorf("create redis mq failed: %w", err))
		}

		logh := slog.NewTextHandler(io.Discard, nil)
		opts := []redismq.OptionFunc{
			redismq.WithContext(ctx),
			redismq.WithLogHandler(logh),
			redismq.WithRedisClient(client),
			redismq.WithNamespaces(cfg.Data.Mq.Redis.Namespaces),
		}

		if value := strings.TrimSpace(cfg.Data.Mq.KeepaliveInterval); value != "" {
			keepalive, err := time.ParseDuration(value)
			if err != nil {
				cancel()
				return nil, errors.Join(cacheClient.Close(), client.Close(), fmt.Errorf("parse data.mq.keepaliveInterval: %w", err))
			}
			opts = append(opts, redismq.WithKeepInterval(keepalive))
		}

		if value := strings.TrimSpace(cfg.Data.Mq.ConsumeTimeout); value != "" {
			consumeTimeout, err := time.ParseDuration(value)
			if err != nil {
				cancel()
				return nil, errors.Join(cacheClient.Close(), client.Close(), fmt.Errorf("parse data.mq.consumeTimeout: %w", err))
			}
			opts = append(opts, redismq.WithConsumeTimeout(consumeTimeout))
		}

		if cfg.Data.Mq.MaxRetry > 0 {
			opts = append(opts, redismq.WithMaxRetry(int(cfg.Data.Mq.MaxRetry)))
		}
		messageQueue = redismq.NewRedisMq(opts...)
		messageQueueClose = client.Close
	default:
		cancel()
		return nil, errors.Join(cacheClient.Close(), fmt.Errorf("unsupported mq driver %q", cfg.Data.Mq.Driver))
	}

	grpcListener, err := net.Listen(cfg.Server.Grpc.Network, cfg.Server.Grpc.Addr)
	if err != nil {
		cancel()
		return nil, errors.Join(cacheClient.Close(), closeMessageQueue(messageQueueClose), fmt.Errorf("create gRPC listener: %w", err))
	}

	httpListener, err := net.Listen(cfg.Server.Http.Network, cfg.Server.Http.Addr)
	if err != nil {
		cancel()
		return nil, errors.Join(cacheClient.Close(), closeMessageQueue(messageQueueClose), grpcListener.Close(), fmt.Errorf("create HTTP listener: %w", err))
	}

	httpTimeout, err := time.ParseDuration(cfg.Server.Http.Timeout)
	if err != nil {
		cancel()
		return nil, errors.Join(cacheClient.Close(), closeMessageQueue(messageQueueClose), grpcListener.Close(), httpListener.Close(), fmt.Errorf("parse server.http.timeout: %w", err))
	}

	repositoryCollection := repository.NewCollection()
	serviceCollection := service.NewCollection(
		service.WithLogHandler(logger.Handler()),
		service.WithEchoRepository(repositoryCollection.Echo),
	)

	return &App{
		ctx:               ctx,
		cache:             cacheClient,
		messageQueue:      messageQueue,
		messageQueueClose: messageQueueClose,
		cancel:            cancel,
		server: handler.New(
			handler.WithLogHandler(logger.Handler()),
			handler.WithServiceCollection(serviceCollection),
			handler.WithGRPCListener(grpcListener),
			handler.WithHTTPListener(httpListener),
			handler.WithAPIPrefix(cfg.Server.Http.ApiPrefix),
			handler.WithHTTPTimeout(httpTimeout),
		),
		logger: logger,
	}, nil
}

func (a *App) Start() error {
	return a.server.Serve(a.ctx)
}

func (a *App) Shutdown(ctx context.Context) error {
	a.shutdownOnce.Do(func() {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		a.logger.Info("server shutting down")
		if a.cancel != nil {
			a.cancel()
		}
		if a.server != nil {
			if err := a.server.Stop(ctx); err != nil {
				a.shutdownErr = err
				return
			}
		}

		var err error
		if a.cache != nil {
			err = errors.Join(err, a.cache.Close())
		}
		if a.messageQueueClose != nil {
			err = errors.Join(err, a.messageQueueClose())
		}
		a.shutdownErr = err
	})

	return a.shutdownErr
}

func closeMessageQueue(close func() error) error {
	if close == nil {
		return nil
	}
	return close()
}
