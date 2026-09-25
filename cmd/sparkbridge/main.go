package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"runtime/pprof"
	"syscall"

	"sparkbridge/internal/config"
	spbgrpc "sparkbridge/internal/grpc"
	"sparkbridge/internal/sparkplug"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "config/config.example.yaml", "path to config file")
	flag.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	cfg, err := config.Load(configPath)
	if err != nil {
		logger.Error("load config", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	mgr, err := sparkplug.NewManager(ctx, cfg, logger)
	if err != nil {
		logger.Error("create manager", "error", err)
		os.Exit(1)
	}
	defer mgr.Close()

	server, err := spbgrpc.NewServer(spbgrpc.Config{SocketPath: cfg.GRPC.SocketPath, SocketPermissions: cfg.GRPC.SocketPermissions}, mgr, logger)
	if err != nil {
		logger.Error("create grpc server", "error", err)
		os.Exit(1)
	}

	if err := server.Start(); err != nil {
		logger.Error("start grpc server", "error", err)
		os.Exit(1)
	}

	go func() {
		<-ctx.Done()
		if p := pprof.Lookup("goroutine"); p != nil {
			_ = p.WriteTo(os.Stderr, 1)
		}
		if err := server.Stop(context.Background()); err != nil && !errors.Is(err, context.Canceled) {
			logger.Error("stop grpc server", "error", err)
		}
	}()

	if err := server.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		logger.Error("grpc server exited", "error", err)
		os.Exit(1)
	}

	fmt.Println("sparkbridge stopped")
}
