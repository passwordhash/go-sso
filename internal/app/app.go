package app

import (
    grpcapp "go-sso/internal/app/grpc"
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

    // TODO: init auth service

    grpcApp := grpcapp.New(log, grpcPort)

    return &App{GRPCSrv: grpcApp}
}
