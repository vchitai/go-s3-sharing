package domain

import "errors"

// Common domain errors
var (
	ErrNotFound     = errors.New("not found")
	ErrUnauthorized = errors.New("unauthorized")
	ErrExpired      = errors.New("expired")
	ErrInvalidPath  = errors.New("invalid path")
	ErrInvalidDate  = errors.New("invalid date")
)
