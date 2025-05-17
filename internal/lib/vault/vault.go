package vault

import (
	"context"
	"fmt"
	"go-sso/internal/services/auth"
	"net/http"
	"time"

	"github.com/hashicorp/vault-client-go"
	"github.com/hashicorp/vault-client-go/schema"
	"go.uber.org/zap"
)

const (
	signingKeyDataKey = "key"

	mountPath   = "secret"
	secretsPath = "/kv/go-sso/clients"
)

type Client struct {
	api *vault.Client
}

// TODO: переместить в storage слой ?

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

func (c *Client) SaveKey(ctx context.Context, appName string, key string) error {
	const op = "vault.SaveKey"

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
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (c *Client) Key(ctx context.Context, appName string) (string, error) {
	const op = "vault.Key"

	appPath := fmt.Sprintf("%s/%s", secretsPath, appName)

	resp, err := c.api.Secrets.KvV2Read(ctx, appPath, vault.WithMountPath(mountPath))
	if err != nil {
		respErr := err.(*vault.ResponseError)
		if respErr != nil {
			if respErr.StatusCode == http.StatusNotFound {
				return "", fmt.Errorf("%w: %v", auth.ErrKeyNotFound, err)
			}
		}

		return "", fmt.Errorf("%s: %w", op, err)
	}

	key, ok := resp.Data.Data[signingKeyDataKey].(string)
	if !ok || key == "" {
		return "nil", fmt.Errorf("%s: %w", op, auth.ErrKeyNotFound)
	}

	return key, nil
}
