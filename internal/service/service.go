package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
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
}

type S3 interface {
	SaveAvatar(ctx context.Context, id string, avatar []byte) (string, error)
}

type Service struct {
	storage Storage
	s3      S3
}

func New(storage Storage, s3 S3) *Service {
	return &Service{
		storage: storage,
		s3:      s3,
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

	hashPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	err = s.storage.SaveUser(
		ctx,
		id,
		email,
		name,
		string(hashPassword),
		about,
		skills,
		avatarUrl,
	)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}
