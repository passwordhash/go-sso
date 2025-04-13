package app

import (
    grpcapp "go-sso/internal/app/grpc"
    "go-sso/internal/services/auth"
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
) *App {
    // TODO: init db

    authService := auth.New(log, nil, nil, nil, 0)

    grpcApp := grpcapp.New(log, authService, grpcPort)

    return &App{GRPCSrv: grpcApp}
}
