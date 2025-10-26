package userServer

import (
	"context"
	userGrpcModels "dormitory-helper-service/generated/proto/user"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserService interface {
	CheckAuthentication(ctx context.Context, token string) (userId int, username string, resultToken string, err error)
}

type Server struct {
	userGrpcModels.UnimplementedUserServiceServer
	service UserService
}

func NewServer(service UserService) *Server {
	return &Server{
		UnimplementedUserServiceServer: userGrpcModels.UnimplementedUserServiceServer{},
		service:                        service,
	}
}

func (s *Server) CheckAuthentication(ctx context.Context, req *userGrpcModels.CheckAuthenticationRequest) (*userGrpcModels.CheckAuthenticationResponse, error) {
	// Вызываем сервис для проверки аутентификации
	userId, username, token, err := s.service.CheckAuthentication(ctx, req.Token)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to check authentication: %v", err)
	}

	// Формируем ответ
	return &userGrpcModels.CheckAuthenticationResponse{
		UserId:   int32(userId),
		Username: username,
		Token:    token,
	}, nil
}
