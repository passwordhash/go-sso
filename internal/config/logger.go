package config

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func SetupLogger(env string) *zap.SugaredLogger {
	var log *zap.Logger
	var err error

	switch env {
	case envLocal, envDevelop:
		log, err = zap.NewDevelopment()
	case envProd:
		traceOpt := zap.AddStacktrace(zapcore.ErrorLevel)
		log, err = zap.NewProduction(traceOpt)
	default:
		panic("unknown environment: " + env)
	}

	if err != nil || log == nil {
		panic(fmt.Sprintf("failed to create logger: %v", err))
	}

	return log.Sugar()
}
