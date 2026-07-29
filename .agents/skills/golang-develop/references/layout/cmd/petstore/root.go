package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"golang-develop-example/internal/petstore/app"
	"golang-develop-example/internal/petstore/config"
	"golang-develop-example/pkg/logging"

	"gitcode.yealink.com/server/devops/yops/ykit/configer"
	runtimex "gitcode.yealink.com/server/devops/yops/ykit/util/runtime"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "petstore",
	Short: "Start App 1",
	Long:  `Start App 1.`,
	Run: func(cmd *cobra.Command, args []string) {
		Run()
	},
}

var rootFlags = struct {
	ConfigFile string
}{}

var cfger *configer.Configer

func init() {
	rootCmd.PersistentFlags().StringVarP(&rootFlags.ConfigFile, "config-file", "c",
		"configs/petstore/config.yml", "specify the config file.")

	cfger = config.NewConfiger(rootCmd)
}

func Run() {
	// Create context
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Load configuration
	c := &config.Config{}
	if err := cfger.Load(c, rootFlags.ConfigFile); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	// Setup log handler
	logHandlers, err := logging.NewLogHandlerSet(c.Log)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	// Preparing
	logger := slog.New(logHandlers.Get(c.Log.Default))
	logger.Debug("Preparing...",
		slog.Any("flags", rootFlags),
		slog.Any("config", c),
		slog.String("pid", fmt.Sprintf("%d", os.Getpid())),
	)

	// Create application
	app, err := app.New(ctx, c, logger)
	if err != nil {
		logger.Error("Create application failed", slog.String("err", err.Error()))
		os.Exit(1)
	}
	defer app.Shutdown(context.Background())

	// Graceful shutdown
	runtimex.SetupGracefulShutdown(ctx, func(sig os.Signal) {
		logger.Info(fmt.Sprintf("receive %s signal, exiting...", sig.String()))
		cancel()
		if err := app.Shutdown(context.Background()); err != nil {
			logger.Error("Shutdown failed", slog.String("err", err.Error()))
			os.Exit(1)
		}
		os.Exit(0)
	})

	// Run application
	logger.Info("Serving...", versionLog)
	if err := app.Start(); err != nil {
		logger.Error("Serving failed", slog.Any("err", err))
		os.Exit(1)
	}
}
