package auth

import (
	"context"
	"errors"
	"go-sso/internal/services/auth"

	vaultlib "go-sso/internal/lib/vault"

	gossov1 "github.com/passwordhash/protos/gen/go/go-sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Auth интерфейс для аутентификации пользователей (сервисная часть).
type Auth interface {
	RegisterNewUser(ctx context.Context,
		email string,
		password string,
	) (userUUID string, err error)

	Login(ctx context.Context,
		email string,
		password string,
		appID int,
	) (token string, err error)

	SigningKey(ctx context.Context, appName string) (key string, err error)
}

type serverAPI struct {
	gossov1.UnimplementedAuthServer
	auth Auth
}

func Register(gRPC *grpc.Server, Vault *vaultlib.Client, auth Auth) {
	gossov1.RegisterAuthServer(gRPC, &serverAPI{
		auth: auth,
	})
}

const (
	emptyValue = 0
)

func (s *serverAPI) Register(ctx context.Context, req *gossov1.RegisterRequest,
) (*gossov1.RegisterResponse, error) {
	if err := validateRegister(req); err != nil {
		return nil, err
	}

	userUUID, err := s.auth.RegisterNewUser(ctx, req.GetEmail(), req.GetPassword())
	if serr := s.handleServiceErr(err); serr != nil {
		return nil, serr
	}

	return &gossov1.RegisterResponse{
		UserUuid: userUUID,
	}, nil
}

func (s *serverAPI) Login(ctx context.Context, req *gossov1.LoginRequest,
) (*gossov1.LoginResponse, error) {
	if err := validateLogin(req); err != nil {
		return nil, err
	}

	token, err := s.auth.Login(ctx, req.GetEmail(), req.GetPassword(), int(req.GetAppId()))
	if serr := s.handleServiceErr(err); serr != nil {
		return nil, serr
	}

	return &gossov1.LoginResponse{
		Token: token,
	}, nil
}

func (s *serverAPI) SigningKey(
	ctx context.Context,
	req *gossov1.SigningKeyRequest,
) (*gossov1.SigningKeyResponse, error) {
	if req.AppName == "" {
		return nil, status.Error(codes.InvalidArgument, "app name is required")
	}

	key, err := s.auth.SigningKey(ctx, req.AppName)
	if err != nil {
		panic("handle errors")
	}

	return &gossov1.SigningKeyResponse{
		SigningKey: key,
	}, nil
}

func (s *serverAPI) handleServiceErr(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, auth.ErrUserExists):
		return status.Error(codes.AlreadyExists, errors.Unwrap(err).Error())
	case errors.Is(err, auth.ErrUserNotFound):
		return status.Error(codes.NotFound, errors.Unwrap(err).Error())
	case errors.Is(err, auth.ErrInvalidCredentials):
		return status.Error(codes.Unauthenticated, errors.Unwrap(err).Error())
	case errors.Is(err, auth.ErrInvalidAppID):
		return status.Error(codes.NotFound, errors.Unwrap(err).Error())
	default:
		return status.Error(codes.Internal, "internal error")
	}
}

func validateLogin(req *gossov1.LoginRequest) error {
	// TODO: add validate lib
	if req.GetEmail() == "" {
		return status.Errorf(codes.InvalidArgument, "email is required")
	}

	if req.GetPassword() == "" {
		return status.Errorf(codes.InvalidArgument, "password is required")
	}

	if req.GetAppId() == emptyValue {
		return status.Errorf(codes.InvalidArgument, "app_id is required")
	}

	return nil
}

func validateRegister(req *gossov1.RegisterRequest) error {
	// TODO: add validate lib
	if req.GetEmail() == "" {
		return status.Errorf(codes.InvalidArgument, "email is required")
	}

	if req.GetPassword() == "" {
		return status.Errorf(codes.InvalidArgument, "password is required")
	}

	return nil
}
