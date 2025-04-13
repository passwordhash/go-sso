package auth

import (
    "context"
    "go-sso/internal/domain/models"
    "go.uber.org/zap"
    "golang.org/x/crypto/bcrypt"
    "time"
)

type Auth struct {
    log          *zap.SugaredLogger
    userSaver    UserSaver
    userProvider UserProvider
    appProvider  AppProvider
    tokenTTL     time.Duration
}

type UserSaver interface {
    SaveUser(ctx context.Context, email string, passHash []byte) (uid int64, err error)
}

type UserProvider interface {
    User(ctx context.Context, email string) (models.User, error)
    IsAdmin(ctx context.Context, email string) (bool, error)
}

type AppProvider interface {
    App(ctx context.Context, appID string) (models.App, error)
}

// New возвращает новый экземпляр сервиса аутентификации.
func New(
        log *zap.SugaredLogger,
        userSaver UserSaver,
        userProvider UserProvider,
        appProvider AppProvider,
        tokenTTL time.Duration,
) *Auth {
    return &Auth{
        log:          log,
        userSaver:    userSaver,
        userProvider: userProvider,
        appProvider:  appProvider,
        tokenTTL:     tokenTTL,
    }
}

// Login проверяет логин и пароль пользователя и возвращает токен.
func (a *Auth) Login(ctx context.Context, email, password string, appID int) (string, error) {
    panic("not implemented")
}

// RegisterNewUser регистрирует нового пользователя и возвращает токен.
// Если пользователь с таким email уже существует, возвращает ошибку.
func (a *Auth) RegisterNewUser(ctx context.Context, email, password string) (int64, error) {
    const op = "auth.RegisterNewUser"

    log := a.log.With("op", op, "email", email)

    log.Infow("registering new user")

    passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        log.Errorw("failed to hash password", "error", err)

        return 0, err
    }

    uid, err := a.userSaver.SaveUser(ctx, email, passHash)
    if err != nil {
        log.Errorw("failed to save user", "error", err)

        return 0, err
    }

    return uid, nil
}

// IsAdmin проверяет, является ли пользователь администратором по id.
func (a *Auth) IsAdmin(ctx context.Context, userID int64) (bool, error) {
    panic("not implemented")
}
