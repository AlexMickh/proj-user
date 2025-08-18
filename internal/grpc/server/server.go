package server

import (
	"context"
	"errors"

	"github.com/AlexMickh/proj-protos/pkg/api/user"
	"github.com/AlexMickh/proj-user/internal/models"
	"github.com/AlexMickh/proj-user/internal/service"
	"github.com/AlexMickh/proj-user/internal/storage"
	"github.com/AlexMickh/proj-user/pkg/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
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
	UserByEmail(ctx context.Context, email string) (models.User, error)
	VerifyEmail(ctx context.Context, id string) error
	UserById(ctx context.Context, id string) (models.User, error)
	UsersBySkills(ctx context.Context, userId string, skills []string) ([]models.User, error)
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
		log.Error("name is empty")
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	if req.GetPassword() == "" {
		log.Error("password is empty")
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}
	if req.GetSkills() == nil {
		log.Error("skills is empty")
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

func (s *Server) GetUserByEmail(ctx context.Context, req *user.GetUserByEmailRequest) (*user.GetUserByEmailResponse, error) {
	const op = "grpc.server.GetUserByEmail"

	log := logger.FromCtx(ctx).With(zap.String("op", op))

	if req.GetEmail() == "" {
		log.Error("email is empty")
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}

	userInfo, err := s.service.UserByEmail(ctx, req.GetEmail())
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Error("user not found", zap.Error(err))
			return nil, status.Error(codes.NotFound, storage.ErrUserNotFound.Error())
		}
		if errors.Is(err, service.ErrEmailNotVerify) {
			log.Error("email not verify", zap.Error(err))
			return nil, status.Error(codes.Unauthenticated, service.ErrEmailNotVerify.Error())
		}
		log.Error("failed to get user", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to get user")
	}

	return &user.GetUserByEmailResponse{
		User: &user.UserType{
			Id:              userInfo.ID,
			Email:           userInfo.Email,
			Name:            userInfo.Name,
			Password:        userInfo.Password,
			About:           &userInfo.About,
			Skills:          userInfo.Skills,
			AvatarUrl:       userInfo.AvatarUrl,
			IsEmailVerified: userInfo.IsEmailVerified,
		},
	}, nil
}

func (s *Server) VerifyEmail(ctx context.Context, req *user.VerifyEmailRequest) (*emptypb.Empty, error) {
	const op = "grpc.server.VerifyEmail"

	log := logger.FromCtx(ctx).With(zap.String("op", op))

	if req.GetId() == "" {
		log.Error("id is empty")
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	err := s.service.VerifyEmail(ctx, req.GetId())
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Error("user not found", zap.Error(err))
			return nil, status.Error(codes.NotFound, storage.ErrUserNotFound.Error())
		}
		log.Error("failed to verify email", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to verify email")
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) GetUserById(ctx context.Context, req *user.GetUserByIdRequest) (*user.GetUserByIdResponse, error) {
	const op = "grpc.server.GetUserById"

	log := logger.FromCtx(ctx).With(zap.String("op", op))

	id := req.GetId()
	if id == "" {
		log.Error("id is empty")
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	userInfo, err := s.service.UserById(ctx, id)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Error("user not found", zap.Error(err))
			return nil, status.Error(codes.NotFound, storage.ErrUserNotFound.Error())
		}
		if errors.Is(err, service.ErrEmailNotVerify) {
			log.Error("email not verify", zap.Error(err))
			return nil, status.Error(codes.Unauthenticated, service.ErrEmailNotVerify.Error())
		}
		log.Error("failed to get user", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to get user")
	}

	return &user.GetUserByIdResponse{
		User: &user.UserType{
			Id:              userInfo.ID,
			Email:           userInfo.Email,
			Name:            userInfo.Name,
			Password:        userInfo.Password,
			About:           &userInfo.About,
			Skills:          userInfo.Skills,
			AvatarUrl:       userInfo.AvatarUrl,
			IsEmailVerified: userInfo.IsEmailVerified,
		},
	}, nil
}

func (s *Server) GetUsersBySkills(ctx context.Context, req *user.GetUsersBySkillsRequest) (*user.GetUsersBySkillsResponse, error) {
	const op = "grpc.server.GetUsersBySkills"

	log := logger.FromCtx(ctx).With(zap.String("op", op))

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Error("failed to get metadata")
		return nil, status.Error(codes.Unauthenticated, "failed to get metadata")
	}

	userId, ok := md["user_id"]
	if !ok {
		log.Error("failed to get user id")
		return nil, status.Error(codes.Unauthenticated, "user id is required")
	}

	skills := req.GetSkills()
	// if skills == nil || len(skills) == 0 {
	// 	log.Error("skills is empty")
	// 	return nil, status.Error(codes.InvalidArgument, "skills is required")
	// }

	usersInfo, err := s.service.UsersBySkills(ctx, userId[0], skills)
	if err != nil {
		log.Error("failed to get users")
		return nil, status.Error(codes.InvalidArgument, "failed to get users")
	}

	users := make([]*user.UserType, 0, len(usersInfo))
	for _, userInfo := range usersInfo {
		user := &user.UserType{
			Id:              userInfo.ID,
			Email:           userInfo.Email,
			Name:            userInfo.Name,
			Password:        userInfo.Password,
			About:           &userInfo.About,
			Skills:          userInfo.Skills,
			AvatarUrl:       userInfo.AvatarUrl,
			IsEmailVerified: userInfo.IsEmailVerified,
		}
		users = append(users, user)
	}

	return &user.GetUsersBySkillsResponse{
		User: users,
	}, nil
}
