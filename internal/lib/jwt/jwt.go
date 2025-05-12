package jwt

import (
	"crypto/rand"
	"encoding/base64"
	"go-sso/internal/domain/models"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TODO: test
func NewToken(user models.User, app models.App, duration time.Duration) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["uuid"] = user.UUID
	claims["email"] = user.Email
	claims["exp"] = time.Now().Add(duration).Unix()
	claims["app_id"] = app.ID

	// TODO: подумать о безопасном хранении секретов
	tokenString, err := token.SignedString([]byte(app.Secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func GenerateHS256Secret() (string, error) {
	bytes := make([]byte, 32) // 256 бит
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(bytes), nil
}
