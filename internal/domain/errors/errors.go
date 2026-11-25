package errors

import "errors"

var (
	ErrUserNotFound             = errors.New("User not found")
	ErrInvalidCredentials       = errors.New("Invalid credentials")
	ErrUserExist                = errors.New("User already exist")
	ErrRefreshTokenNotFound     = errors.New("Refresh token not found")
	ErrRefreshTokenByIDNotFound = errors.New("Refreh token by id not found")
	ErrSecretNotFound           = errors.New("Secrets not found")
	ErrAppNotFound              = errors.New("App not found")
)
