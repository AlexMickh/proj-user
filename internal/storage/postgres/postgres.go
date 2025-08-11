package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/AlexMickh/proj-user/internal/models"
	"github.com/AlexMickh/proj-user/internal/storage"
	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Postgres interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type Storage struct {
	db   Postgres
	psql sq.StatementBuilderType
}

func New(db Postgres) *Storage {
	return &Storage{
		db:   db,
		psql: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
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

	query, args, err := s.psql.Insert("users").
		Columns("id", "email", "name", "password", "about", "skills", "avatar_url").
		Values(id, email, name, password, about, skills, avatarUrl).
		ToSql()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = s.db.Exec(ctx, query, args...)
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

	query, args, err := s.psql.Select("id", "email", "name", "password", "about", "skills", "avatar_url", "is_email_verified").
		From("users").
		Where("email = ?", email).
		ToSql()
	if err != nil {
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	var user models.User
	err = s.db.QueryRow(ctx, query, args...).Scan(
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

func (s *Storage) VerifyEmail(ctx context.Context, id string) (models.User, error) {
	const op = "storage.postgres.VerifyEmail"

	query, args, err := s.psql.Update("users").
		Set("is_email_verified", true).
		Where("id = ?", id).
		Suffix("RETURNING id, email, name, password, about, skills, avatar_url, is_email_verified").
		ToSql()
	if err != nil {
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	var user models.User
	err = s.db.QueryRow(ctx, query, args...).Scan(
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

func (s *Storage) UserById(ctx context.Context, id string) (models.User, error) {
	const op = "storage.postgres.UserById"

	query, args, err := s.psql.Select("id", "email", "name", "password", "about", "skills", "avatar_url", "is_email_verified").
		From("users").
		Where("id = ?", id).
		ToSql()
	if err != nil {
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	var user models.User
	err = s.db.QueryRow(ctx, query, args...).Scan(
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

func (s *Storage) UsersBySkills(ctx context.Context, skills []string) ([]models.User, error) {
	const op = "storage.postgres.UsersBySkills"

	query, args, err := s.psql.Select("id", "email", "name", "password", "about", "skills", "avatar_url", "is_email_verified").
		From("users").
		Where("skills && ? AND is_email_verified = ?", skills, true).
		Limit(10).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		rows.Scan(
			&user.ID,
			&user.Email,
			&user.Name,
			&user.Password,
			&user.About,
			&user.Skills,
			&user.AvatarUrl,
			&user.IsEmailVerified,
		)
		users = append(users, user)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return users, nil
}
