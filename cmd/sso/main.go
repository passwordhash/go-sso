package main

import (
	"context"
	"go-sso/internal/app"
	"go-sso/internal/config"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx := context.Background()

	cfg := config.MustLoad()

	log := config.SetupLogger(cfg.Env)

	log.Info("starting SSO application...")
	log.Debugw("with config", "config", cfg)

	application := app.New(ctx, log, cfg)

	go application.GRPCSrv.MustRun()

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	sign := <-stop

	log.Infow("received signal", "signal", sign)

	application.GRPCSrv.Stop()

	log.Infow("stopped SSO application")
}
