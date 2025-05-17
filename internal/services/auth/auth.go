package auth

import (
	"context"
	"errors"
	"fmt"
	"go-sso/internal/domain/models"
	"go-sso/internal/lib/jwt"
	"go-sso/internal/storage"
	"time"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type Auth struct {
	log *zap.SugaredLogger

	userSaver          UserSaver
	userProvider       UserProvider
	appProvider        AppProvider
	signingKeySaver    SigningKeySaver
	signingKeyProvider SigningKeyProvider

	tokenTTL time.Duration
}

type UserSaver interface {
	SaveUser(ctx context.Context, email string, passHash []byte) (uuid string, err error)
}

type UserProvider interface {
	User(ctx context.Context, email string) (models.User, error)
}

type AppProvider interface {
	App(ctx context.Context, appID int) (models.App, error)
}

type SigningKeySaver interface {
	SaveKey(ctx context.Context, appName string, key string) error
}

type SigningKeyProvider interface {
	Key(ctx context.Context, appName string) (string, error)
}

var (
	ErrKeyNotFound = errors.New("key not found")
)

// Ошибки, которые могут возникнуть при работе с сервисом аутентификации.
var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserExists         = errors.New("user already exists")
	ErrInvalidAppID       = errors.New("invalid app id")
)

// New возвращает новый экземпляр сервиса аутентификации.
func New(
	log *zap.SugaredLogger,
	userSaver UserSaver,
	userProvider UserProvider,
	appProvider AppProvider,
	signingKeySaver SigningKeySaver,
	signingKeyProvider SigningKeyProvider,
	tokenTTL time.Duration,
) *Auth {
	return &Auth{
		log: log,

		userSaver:          userSaver,
		userProvider:       userProvider,
		appProvider:        appProvider,
		signingKeySaver:    signingKeySaver,
		signingKeyProvider: signingKeyProvider,

		tokenTTL: tokenTTL,
	}
}

// Login проверяет логин и пароль пользователя и возвращает токен.
func (a *Auth) Login(ctx context.Context, email, password string, appName string) (string, error) {
	const op = "auth.Login"

	log := a.log.With("op", op, "email", email, "appName", appName)

	log.Infow("logging in user")

	user, err := a.userProvider.User(ctx, email)
	if err := handleStorageErr(log, err, op); err != nil {
		return "", err
	}
	if err != nil {
		return "", handleInternalErr(log, "failed to get user", op, err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		a.log.Infow("invalid credentials", "error", err)

		return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	log.Infow("user logged in", "userID", user.UUID)

	secret, err := a.signingKeyProvider.Key(ctx, appName)
	if errors.Is(err, ErrKeyNotFound) {
		log.Infow("key not found, generating new key", "error", err)

		secret, err = jwt.GenerateHS256Secret()
		if err != nil {
			log.Errorw("failed to generate signing key", "error", err)

			return "", fmt.Errorf("%s: %w", op, err)
		}

		if saveErr := a.signingKeySaver.SaveKey(ctx, appName, secret); saveErr != nil {
			log.Errorw("failed to save signing key", "error", saveErr)

			return "", fmt.Errorf("%s: %w", op, saveErr)
		}
	} else if err != nil {
		log.Errorw("failed to get signing key", "error", err)

		return "", fmt.Errorf("%s: %w", op, err)
	}

	token, err := jwt.NewToken(user, secret, a.tokenTTL)
	if err != nil {
		return "", handleInternalErr(log, "failed to create token", op, err)
	}

	return token, nil
}

// RegisterNewUser регистрирует нового пользователя и возвращает токен.
// Если пользователь с таким email уже существует, возвращает ошибку.
func (a *Auth) RegisterNewUser(ctx context.Context, email, password string) (string, error) {
	const op = "auth.RegisterNewUser"

	log := a.log.With("op", op, "email", email)
	log.Infow("registering new user")

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", handleInternalErr(log, "failed to hash password", op, err)
	}

	userUUID, err := a.userSaver.SaveUser(ctx, email, passHash)
	if sterr := handleStorageErr(log, err, op); sterr != nil {
		return "", sterr
	}
	if err != nil {
		return "", handleInternalErr(log, "failed to save user", op, err)
	}

	log.Infow("user registered", "userUUID", userUUID)

	return userUUID, nil
}

// SigningKey возвращает ключ подписи для приложения с заданным именем.
func (a *Auth) SigningKey(ctx context.Context, appName string) (string, error) {
	const op = "auth.SigningKey"

	log := a.log.With("op", op, "appName", appName)
	log.Infow("getting signing key")

	key, err := a.signingKeyProvider.Key(ctx, appName)
	// TODO: refactor flat
	if errors.Is(err, ErrKeyNotFound) {
		log.Infow("key not found, generating new key", "error", err)

		secret, err := jwt.GenerateHS256Secret()
		if err != nil {
			log.Errorw("failed to generate signing key", "error", err)
			return "", fmt.Errorf("%s: %w", op, err)
		}

		if saveErr := a.signingKeySaver.SaveKey(ctx, appName, secret); saveErr != nil {
			log.Errorw("failed to save signing key", "error", saveErr)
			return "", fmt.Errorf("%s: %w", op, saveErr)
		}

		return secret, nil
	}
	if err != nil {
		log.Errorw("failed to get signing key", "error", err)
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return key, nil
}

// handleStorageErr обрабатывает ошибки, возвращаемые хранилищем и логгирует их.
// Если ошибка не является ошибкой хранилища, возвращает nil.
func handleStorageErr(log *zap.SugaredLogger, err error, op string) error {
	switch {
	case errors.Is(err, storage.ErrUserExists):
		log.Warnw("user already exists", "error", err.Error())
		return fmt.Errorf("%s: %w", op, ErrUserExists)

	case errors.Is(err, storage.ErrUserNotFound):
		log.Infow("user not found", "error", err)
		return fmt.Errorf("%s: %w", op, ErrInvalidCredentials)

	case errors.Is(err, storage.ErrAppNotFound):
		log.Infow("app not found", "error", err)
		return fmt.Errorf("%s: %w", op, ErrInvalidAppID)

	default:
		return nil // значит это не "известная" ошибка
	}
}

func handleInternalErr(log *zap.SugaredLogger, msg, op string, err error) error {
	log.Errorw(msg, "error", err)
	return fmt.Errorf("%s: %w", op, err)
}
