package main

import (
    "go-sso/internal/app"
    "go-sso/internal/config"
    "os"
    "os/signal"
    "syscall"
)

func main() {
    cfg := config.MustLoad()

    log := config.SetupLogger(cfg.Env)

    log.Infow("starting SSO application...", "config", cfg)

    application := app.New(log, cfg.GRPC.Port, cfg.TokenTTL)

    go application.GRPCSrv.MustRun()

    // Graceful shutdown
    stop := make(chan os.Signal, 1)
    signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

    sign := <-stop

    log.Infow("received signal", "signal", sign)

    application.GRPCSrv.Stop()

    log.Infow("stopped SSO application")
}
