package server

import (
	"context"
	"errors"

	"github.com/AlexMickh/proj-protos/pkg/api/user"
	"github.com/AlexMickh/proj-user/internal/storage"
	"github.com/AlexMickh/proj-user/pkg/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service interface {
	CreateUser(
		ctx context.Context,
		email string,
		name string,
		password string,
		about string,
		skills []string,
		avatar []byte,
	) (string, error)
}

type Server struct {
	user.UnimplementedUserServer
	service Service
}

func New(service Service) *Server {
	return &Server{
		service: service,
	}
}

func (s *Server) CreateUser(ctx context.Context, req *user.CreateUserRequest) (*user.CreateUserResponse, error) {
	const op = "grpc.server.CreateUser"

	log := logger.FromCtx(ctx).With(zap.String("op", op))

	if req.GetEmail() == "" {
		log.Error("email is empty")
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}
	if req.GetName() == "" {
		log.Error("email is empty")
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	if req.GetPassword() == "" {
		log.Error("email is empty")
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}
	if req.GetSkills() == nil {
		log.Error("email is empty")
		return nil, status.Error(codes.InvalidArgument, "skills is required")
	}

	id, err := s.service.CreateUser(
		ctx,
		req.GetEmail(),
		req.GetName(),
		req.GetPassword(),
		req.GetAbout(),
		req.GetSkills(),
		req.GetAvatar(),
	)
	if err != nil {
		if errors.Is(err, storage.ErrUserAlreadyExists) {
			log.Error("user already exists")
			return nil, status.Error(codes.InvalidArgument, storage.ErrUserAlreadyExists.Error())
		}
		if errors.Is(err, storage.ErrInvalidSkills) {
			log.Error("skill not in the skills list")
			return nil, status.Error(codes.InvalidArgument, storage.ErrInvalidSkills.Error())
		}
		log.Error("failed to create user", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to create user")
	}

	return &user.CreateUserResponse{
		Id: id,
	}, nil
}
