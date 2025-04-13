package main

import (
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

    log.Infow("SSO service started", "config", cfg)
}
