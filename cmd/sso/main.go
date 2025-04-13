package main

import (
    "go-sso/internal/app"
    "go-sso/internal/config"
)

const (
    envLocal   = "local"
    envDevelop = "dev"
    envProd    = "prod"
)

func main() {
    cfg := config.MustLoad()

    log := config.SetupLogger(cfg.Env)

    log.Infow("starting SSO application...", "config", cfg)

    application := app.New(log, cfg.GRPC.Port, cfg.TokenTTL)

    application.GRPCSrv.MustRun()
}
