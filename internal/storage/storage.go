package storage

import "errors"

var (
	ErrUserNotFound       = errors.New("User not found")
	ErrInvalidCredentials = errors.New("Invalid credentials")
	ErrUserExist          = errors.New("User already exist")
)
