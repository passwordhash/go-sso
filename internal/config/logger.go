package config

import (
    "fmt"
    "go.uber.org/zap"
)

func SetupLogger(env string) *zap.SugaredLogger {
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
