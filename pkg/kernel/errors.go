package kernel

import "errors"

var (
	ErrNotFound         = errors.New("kernel: resource not found")
	ErrProviderNotFound = errors.New("kernel: provider not found")
	ErrNotImplemented   = errors.New("kernel: operation not implemented")
)
