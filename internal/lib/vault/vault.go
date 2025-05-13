package vault

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/vault-client-go"
	"github.com/hashicorp/vault-client-go/schema"
	"go.uber.org/zap"
)

const (
	signingKeyDataKey = "key"

	mountPath   = "secret"
	secretsPath = "/secret/go-sso/clients"
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
	// Для прода использовать approle
	if err := c.SetToken(token); err != nil {
		log.Fatalw("failed to authenticate with vault", zap.Error(err))
	}

	return &Client{
		api: c,
	}
}

func (c *Client) SaveKey(ctx context.Context, appName string, key []byte) error {
	appPath := fmt.Sprintf("%s/%s", secretsPath, appName)

	secret := map[string]interface{}{
		signingKeyDataKey: key,
	}

	_, err := c.api.Secrets.KvV2Write(ctx,
		appPath,
		schema.KvV2WriteRequest{
			Data: secret,
		},
		vault.WithMountPath(mountPath))
	if err != nil {
		return fmt.Errorf("failed to write secret to vault: %w", err)
	}

	return nil
}

func (c *Client) Key(ctx context.Context, appName string) ([]byte, error) {
	appPath := fmt.Sprintf("%s/%s", secretsPath, appName)

	resp, err := c.api.Secrets.KvV2Read(ctx, appPath, vault.WithMountPath(mountPath))
	if err != nil {
		return nil, fmt.Errorf("failed to read secret from vault: %w", err)
	}

	key, ok := resp.Data.Data[signingKeyDataKey]
	if !ok {
		return nil, fmt.Errorf("key not found in secret: %s", appPath)
	}

	return key.([]byte), nil
}
