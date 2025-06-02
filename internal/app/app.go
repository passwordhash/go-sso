package app

import (
	"context"
	grpcapp "go-sso/internal/app/grpc"
	"go-sso/internal/config"
	vaultlib "go-sso/internal/lib/vault"
	"go-sso/internal/services/auth"
	"go-sso/internal/storage/postgres"

	"go.uber.org/zap"
)

type App struct {
	GRPCSrv *grpcapp.App
}

func New(
	ctx context.Context,
	log *zap.SugaredLogger,
	cfg *config.Config,
) *App {
	storage, err := postgres.New(cfg.PSQL.DSN())
	if err != nil {
		log.Fatalw("failed to connect to PostgreSQL", "error", err)
	}

	vaultClient := vaultlib.New(ctx,
		log,
		cfg.Vault.Addr,
		cfg.Vault.Token,
		cfg.Vault.Timeout,
	)

	authService := auth.New(log, storage, storage, storage, vaultClient, vaultClient, cfg.TokenTTL)

	grpcApp := grpcapp.New(log,
		cfg.AppServiceName,
		vaultClient,
		authService,
		cfg.GRPC.Port,
	)

	return &App{
		GRPCSrv: grpcApp,
	}
}
