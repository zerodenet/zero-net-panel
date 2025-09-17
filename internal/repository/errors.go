package repository

import "errors"

var (
	ErrNotFound        = errors.New("repository: resource not found")
	ErrConflict        = errors.New("repository: conflict")
	ErrInvalidArgument = errors.New("repository: invalid argument")
	ErrForbidden       = errors.New("repository: forbidden")
)
