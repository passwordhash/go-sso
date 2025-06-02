package grpcapp

import (
	"fmt"
	authgrpc "go-sso/internal/grpc/auth"
	"net"

	vaultlib "go-sso/internal/lib/vault"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

type App struct {
	log        *zap.SugaredLogger
	gRPCServer *grpc.Server
	port       int
}

// New создает новый экземпляр gRPC сервера.
func New(
	log *zap.SugaredLogger,
	appServiceName string,
	vaultClient *vaultlib.Client,
	authService authgrpc.Auth,
	port int,
) *App {
	gRPCServer := grpc.NewServer()

	authgrpc.Register(gRPCServer, vaultClient, authService)

	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(gRPCServer, healthServer)
	healthServer.SetServingStatus(appServiceName, grpc_health_v1.HealthCheckResponse_SERVING)

	reflection.Register(gRPCServer)

	return &App{
		log:        log,
		gRPCServer: gRPCServer,
		port:       port,
	}
}

// MustRun запускает gRPC сервер и вызывает панику в случае ошибки.
func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		a.log.Panicw("failed to run gRPC server", "err", err)
	}
}

// Run запускает gRPC сервер и слушает указанный порт.
func (a *App) Run() error {
	const op = "grpcapp.Run"

	log := a.log.With(zap.String("op", op))

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Infow("gRPC server started", "port", l.Addr().String())

	if err := a.gRPCServer.Serve(l); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// Stop останавливает gRPC сервер.
func (a *App) Stop() {
	const op = "grpcapp.Stop"

	log := a.log.With(zap.String("op", op))

	log.Infow("stopping gRPC server")

	a.gRPCServer.GracefulStop()

	log.Infow("gracefully stopped gRPC server")
}
