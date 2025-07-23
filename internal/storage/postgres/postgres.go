package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/AlexMickh/proj-user/internal/models"
	"github.com/AlexMickh/proj-user/internal/storage"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Postgres interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type Storage struct {
	db Postgres
}

func New(db Postgres) *Storage {
	return &Storage{
		db: db,
	}
}

func (s *Storage) SaveUser(
	ctx context.Context,
	id string,
	email string,
	name string,
	password string,
	about string,
	skills []string,
	avatarUrl string,
) error {
	const op = "storage.postgres.SaveUser"

	query := `INSERT INTO users
			(id, email, name, password, about, skills, avatar_url)
			VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := s.db.Exec(ctx, query, id, email, name, password, about, skills, avatarUrl)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "22P02" {
				return fmt.Errorf("%s: %w", op, storage.ErrInvalidSkills)
			}
			if pgErr.Code == "23505" {
				return fmt.Errorf("%s: %w", op, storage.ErrUserAlreadyExists)
			}
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) UserByEmail(ctx context.Context, email string) (models.User, error) {
	const op = "storage.postgres.UserByEmail"

	query := `SELECT id, email, name, password, about, skills, avatar_url, is_email_verified
			FROM users
			WHERE email = $1`
	var user models.User
	err := s.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.Password,
		&user.About,
		&user.Skills,
		&user.AvatarUrl,
		&user.IsEmailVerified,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

func (s *Storage) VerifyEmail(ctx context.Context, email string) (models.User, error) {
	const op = "storage.postgres.VerifyEmail"

	query := `UPDATE users 
			  SET is_email_verified = true 
			  WHERE email = $1
			  RETURNING id, email, name, password, about, skills, avatar_url, is_email_verified`
	var user models.User
	err := s.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.Password,
		&user.About,
		&user.Skills,
		&user.AvatarUrl,
		&user.IsEmailVerified,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}
