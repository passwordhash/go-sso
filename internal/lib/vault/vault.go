package vault

import (
	"context"
	"time"

	"github.com/hashicorp/vault-client-go"
	"go.uber.org/zap"
)

type Client struct {
	api *vault.Client
}

// Создать новый клиент Vault
func New(
	ctx context.Context,
	log *zap.SugaredLogger,
	addr string,
	token string,
	timeout time.Duration,
) *Client {
	c, err := vault.New(
		vault.WithAddress(addr),
		vault.WithRequestTimeout(timeout),
	)
	if err != nil {
		log.Fatalw("failed to create vault client", zap.Error(err))
	}

	// authenticate with a root token (insecure)
	if err := c.SetToken(token); err != nil {
		log.Fatalw("failed to authenticate with vault", zap.Error(err))
	}

	return &Client{
		api: c,
	}
}
