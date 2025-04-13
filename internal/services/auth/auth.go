package auth

import (
    "context"
    "errors"
    "fmt"
    "go-sso/internal/domain/models"
    "go-sso/internal/lib/jwt"
    "go-sso/internal/storage"
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
    IsAdmin(ctx context.Context, uid int64) (bool, error)
}

type AppProvider interface {
    App(ctx context.Context, appID int) (models.App, error)
}

var (
    ErrInvalidCredentials = errors.New("invalid credentials")
    ErrUserNotFound       = errors.New("user not found")
    ErrUserExists         = errors.New("user already exists")
    ErrAppNotFound        = errors.New("app not found")
)

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
    const op = "auth.Login"

    log := a.log.With("op", op, "email", email, "appID", appID)

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

    app, err := a.appProvider.App(ctx, appID)
    if sterr := handleStorageErr(log, err, op); sterr != nil {
        return "", sterr
    }
    if err != nil {
        return "", handleInternalErr(log, "failed to get app", op, err)
    }

    log.Infow("user logged in", "userID", user.ID, "appID", app.ID)

    token, err := jwt.NewToken(user, app, a.tokenTTL)
    if sterr := handleStorageErr(log, err, op); sterr != nil {
        return "", sterr
    }
    if err != nil {
        return "", handleInternalErr(log, "failed to create token", op, err)
    }

    return token, nil
}

// RegisterNewUser регистрирует нового пользователя и возвращает токен.
// Если пользователь с таким email уже существует, возвращает ошибку.
func (a *Auth) RegisterNewUser(ctx context.Context, email, password string) (int64, error) {
    const op = "auth.RegisterNewUser"

    log := a.log.With("op", op, "email", email)

    log.Infow("registering new user")

    passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return 0, handleInternalErr(log, "failed to hash password", op, err)
    }

    uid, err := a.userSaver.SaveUser(ctx, email, passHash)
    if sterr := handleStorageErr(log, err, op); sterr != nil {
        return 0, sterr
    }
    if err != nil {
        return 0, handleInternalErr(log, "failed to save user", op, err)
    }

    return uid, nil
}

// IsAdmin проверяет, является ли пользователь администратором по id.
func (a *Auth) IsAdmin(ctx context.Context, userID int64) (bool, error) {
    const op = "auth.IsAdmin"

    log := a.log.With("op", op, "userID", userID)

    log.Infow("checking if user is admin")

    isAdmin, err := a.userProvider.IsAdmin(ctx, userID)
    if sterr := handleStorageErr(log, err, op); sterr != nil {
        return false, sterr
    }
    if err != nil {
        return false, handleInternalErr(log, "failed to check if user is admin", op, err)
    }

    log.Infow("user is admin", "isAdmin", isAdmin)

    return isAdmin, nil
}

// handleStorageErr обрабатывает ошибки, возвращаемые хранилищем и логгирует их.
// Если ошибка не является ошибкой хранилища, возвращает nil.
func handleStorageErr(log *zap.SugaredLogger, err error, op string) error {
    switch {
    case errors.Is(err, storage.ErrUserExists):
        log.Warnw("user already exists", "error", err)
        return fmt.Errorf("%s: %w", op, ErrUserExists)

    case errors.Is(err, storage.ErrUserNotFound):
        log.Infow("user not found", "error", err)
        return fmt.Errorf("%s: %w", op, ErrUserNotFound)

    case errors.Is(err, storage.ErrAppNotFound):
        log.Infow("app not found", "error", err)
        return fmt.Errorf("%s: %w", op, ErrAppNotFound)

    default:
        return nil // значит это не "известная" ошибка
    }
}

func handleInternalErr(log *zap.SugaredLogger, msg, op string, err error) error {
    log.Errorw(msg, "error", err)
    return fmt.Errorf("%s: %w", op, err)
}
