package storage

import "errors"

var (
	ErrInvalidSkills     = errors.New("skill not in the skills list")
	ErrUserAlreadyExists = errors.New("user already exists")
)
