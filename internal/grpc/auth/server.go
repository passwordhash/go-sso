package auth

import (
    "context"
    gossov1 "github.com/passwordhash/protos/gen/go/go-sso"
    "google.golang.org/grpc"
)

type serverAPI struct {
    gossov1.UnimplementedAuthServer
}

func Register(gRPC *grpc.Server) {
    gossov1.RegisterAuthServer(gRPC, &serverAPI{})
}

func (s *serverAPI) Login(ctx context.Context, req *gossov1.LoginRequest,
) (*gossov1.LoginResponse, error) {
    return &gossov1.LoginResponse{
        Token: req.GetEmail(),
    }, nil
}

func (s *serverAPI) Register(ctx context.Context, req *gossov1.RegisterRequest,
) (*gossov1.RegisterResponse, error) {
    panic("implement me")
}

func (s *serverAPI) IsAdmin(ctx context.Context, req *gossov1.IsAdminRequest,
) (*gossov1.IsAdminResponse, error) {
    panic("implement me")
}
