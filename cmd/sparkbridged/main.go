package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"sparkbridge/pkg/config"
	"sparkbridge/pkg/manager"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "configs/sparkbridge.yaml", "path to config file")
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

	mgr := manager.New()
	defer mgr.Close()
	for _, bridgeCfg := range cfg.Bridges {
		if _, err := mgr.Get(ctx, bridgeCfg); err != nil {
			logger.Error("create bridge", "error", err)
			os.Exit(1)
		}
	}

	<-ctx.Done()
	logger.Info("sparkbridge stopped")
}
