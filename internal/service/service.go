package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/AlexMickh/proj-user/internal/models"
	"github.com/google/uuid"
)

type Storage interface {
	SaveUser(
		ctx context.Context,
		id string,
		email string,
		name string,
		password string,
		about string,
		skills []string,
		avatarUrl string,
	) error
	UserByEmail(ctx context.Context, email string) (models.User, error)
	VerifyEmail(ctx context.Context, id string) (models.User, error)
	UserById(ctx context.Context, id string) (models.User, error)
	UsersBySkills(ctx context.Context, userId string, skills []string) ([]models.User, error)
	RandomUsers(ctx context.Context, userId string) ([]models.User, error)
}

type S3 interface {
	SaveAvatar(ctx context.Context, id string, avatar []byte) (string, error)
}

type Cash interface {
	SaveUser(ctx context.Context, user models.User) error
	UserByEmail(ctx context.Context, email string) (models.User, error)
	UpdateUser(ctx context.Context, user models.User) error
	UserById(ctx context.Context, id string) (models.User, error)
}

type Service struct {
	storage Storage
	s3      S3
	cash    Cash
}

var ErrEmailNotVerify = errors.New("email is not verify")

func New(storage Storage, s3 S3, cash Cash) *Service {
	return &Service{
		storage: storage,
		s3:      s3,
		cash:    cash,
	}
}

func (s *Service) CreateUser(
	ctx context.Context,
	email string,
	name string,
	password string,
	about string,
	skills []string,
	avatar []byte,
) (string, error) {
	const op = "service.CreateUser"

	id := uuid.NewString()

	avatarUrl, err := s.s3.SaveAvatar(ctx, id, avatar)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	err = s.storage.SaveUser(
		ctx,
		id,
		email,
		name,
		password,
		about,
		skills,
		avatarUrl,
	)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (s *Service) UserByEmail(ctx context.Context, email string) (models.User, error) {
	const op = "service.UserByEmail"

	user, err := s.cash.UserByEmail(ctx, email)
	if err == nil {
		if !user.IsEmailVerified {
			return models.User{}, fmt.Errorf("%s: %w", op, ErrEmailNotVerify)
		}
		return user, nil
	}

	user, err = s.storage.UserByEmail(ctx, email)
	if err != nil {
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}
	if !user.IsEmailVerified {
		return models.User{}, fmt.Errorf("%s: %w", op, ErrEmailNotVerify)
	}

	err = s.cash.SaveUser(ctx, user)
	if err != nil {
		return user, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

func (s *Service) VerifyEmail(ctx context.Context, id string) error {
	const op = "service.VerifyEmail"

	user, err := s.storage.VerifyEmail(ctx, id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	err = s.cash.UpdateUser(ctx, user)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Service) UserById(ctx context.Context, id string) (models.User, error) {
	const op = "service.UserById"

	user, err := s.cash.UserById(ctx, id)
	if err == nil {
		if !user.IsEmailVerified {
			return models.User{}, fmt.Errorf("%s: %w", op, ErrEmailNotVerify)
		}
		fmt.Println("from cash")
		return user, nil
	}

	user, err = s.storage.UserById(ctx, id)
	if err != nil {
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}
	if !user.IsEmailVerified {
		return models.User{}, fmt.Errorf("%s: %w", op, ErrEmailNotVerify)
	}

	err = s.cash.SaveUser(ctx, user)
	if err != nil {
		return user, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

func (s *Service) UsersBySkills(ctx context.Context, userId string, skills []string) ([]models.User, error) {
	const op = "service.UsersBySkills"

	if skills != nil {
		users, err := s.storage.UsersBySkills(ctx, userId, skills)
		if err != nil {
			return nil, err
		}

		return users, nil
	}

	users, err := s.storage.RandomUsers(ctx, userId)
	if err != nil {
		return nil, err
	}

	return users, nil
}
