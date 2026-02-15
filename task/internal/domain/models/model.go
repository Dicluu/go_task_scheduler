package models

import "errors"

var (
	ErrCannotViewRecord   = errors.New("you are unauthorized to view this record")
	ErrCannotUpdateRecord = errors.New("you are unauthorized to update this record")
	ErrCannotDeleteRecord = errors.New("you are unauthorized to delete this record")
	ErrUnauthorized       = errors.New("user not unauthorized")
)
