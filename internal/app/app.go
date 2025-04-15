package app

import (
    grpcapp "go-sso/internal/app/grpc"
    "go-sso/internal/services/auth"
    "go-sso/internal/storage/postgres"
    "go.uber.org/zap"
    "time"
)

type App struct {
    GRPCSrv *grpcapp.App
}

func New(
        log *zap.SugaredLogger,
        grpcPort int,
        tokenTTL time.Duration,
        psqlConn string,
) *App {
    storage, err := postgres.New(psqlConn)
    if err != nil {
        log.Fatalw("failed to connect to PostgreSQL", "error", err)
    }

    authService := auth.New(log, storage, storage, storage, tokenTTL)

    grpcApp := grpcapp.New(log, authService, grpcPort)

    return &App{GRPCSrv: grpcApp}
}
