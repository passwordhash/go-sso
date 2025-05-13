package suite

import (
	"context"
	"go-sso/internal/config"
	"strconv"
	"testing"

	gossov1 "github.com/passwordhash/protos/gen/go/go-sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Suite struct {
	*testing.T
	Cfg        *config.Config
	AuthClient gossov1.AuthClient
}

func New(t *testing.T) (context.Context, *Suite) {
	t.Helper() // Помещаем вызов в функцию, чтобы пометить его как вспомогательный
	t.Parallel()

	// TODO: для тестов в CI/CD нужно использовать переменные окружения
	cfg := config.MustLoadByPath("../config/test.yml")

	ctx, cancelCtx := context.WithTimeout(context.Background(), cfg.GRPC.Timeout)

	t.Cleanup(func() {
		t.Helper()
		cancelCtx()
	})

	conn, err := grpc.NewClient(
		grpcAddr(cfg),
		grpc.WithTransportCredentials(insecure.NewCredentials())) // используем небезопасные соединения для тестов
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}

	return ctx, &Suite{
		T:          t,
		Cfg:        cfg,
		AuthClient: gossov1.NewAuthClient(conn),
	}
}

func grpcAddr(cfg *config.Config) string {
	return cfg.GRPC.Host + ":" + strconv.Itoa(cfg.GRPC.Port)
}
