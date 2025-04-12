package main

import (
    "fmt"
    "go-sso/internal/config"
    "go.uber.org/zap"
)

const (
    envLocal   = "local"
    envDevelop = "dev"
    envProd    = "prod"
)

func main() {
    cfg := config.MustLoad()

    log := setupLogger(cfg.Env)

    log.Infow("SSO service started",
        "env", cfg.Env,
        "config", cfg,
    )
}

func setupLogger(env string) *zap.SugaredLogger {
    var log *zap.Logger
    var err error

    switch env {
    case envLocal, envDevelop:
        log, err = zap.NewDevelopment()
    case envProd:
        log, err = zap.NewProduction()
    default:
        panic("unknown environment: " + env)
    }

    if err != nil || log == nil {
        panic(fmt.Sprintf("failed to create logger: %v", err))
    }

    return log.Sugar()
}
