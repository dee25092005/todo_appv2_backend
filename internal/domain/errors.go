package domain

import "errors"

var (
	ErrNotFound      = errors.New("requested resource not found")
	ErrUnauthorized  = errors.New("unauthorized access")
	ErrInvalidInput  = errors.New("invalid input data")
	ErrInternal      = errors.New("internal error")
	ErrAlreadyExists = errors.New("resource already exists")
)
