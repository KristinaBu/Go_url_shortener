package domain

import "errors"

var (
	ErrNotFound          = errors.New("link not found")
	ErrURLAlreadyExists  = errors.New("original URL already exists")
	ErrCodeAlreadyExists = errors.New("short code already exists")
	ErrInvalidURL        = errors.New("invalid URL")
)
