package usecase

import "errors"

var (
	ErrTokenUsed            = errors.New("token is unavailable")
	ErrWrongLoginOrPassword = errors.New("failed to login. Incorrect login or password")
	ErrUserAlreadyExists    = errors.New("user with current email is already exists")
)
