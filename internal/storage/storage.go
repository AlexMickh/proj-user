package storage

import "errors"

var (
	ErrInvalidSkills     = errors.New("skill not in the skills list")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrUserNotFound      = errors.New("user not found")
)
