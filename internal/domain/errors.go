package domain

import "errors"

var (
	ErrNotFound          = errors.New("not found")
	ErrConflict          = errors.New("conflict")
	ErrInvalidCredential = errors.New("invalid credentials")
	ErrUpstream          = errors.New("upstream service error")
)
