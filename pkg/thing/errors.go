package thing

import "errors"

// Domain-specific errors for the Thing service
var (
	ErrNotFound          = errors.New("thing not found")
	ErrAlreadyExists     = errors.New("thing already exists")
	ErrTypeThingNotFound = errors.New("type thing not found")
	ErrUnauthorized      = errors.New("unauthorized")
	ErrInvalidInput      = errors.New("invalid input")
	ErrNotOwner          = errors.New("user is not the owner")
	ErrAdminRequired     = errors.New("admin privileges required")
)
