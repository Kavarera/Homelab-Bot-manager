package domain

import "errors"

var (
	// ErrUnauthorized is returned when a user does not have permission to access the bot.
	ErrUnauthorized = errors.New("unauthorized access")

	// ErrInvalidCommand is returned when a command format is invalid.
	ErrInvalidCommand = errors.New("invalid command format")

	// ErrInvalidSecurityKeyword is returned when the security keyword fails validation.
	ErrInvalidSecurityKeyword = errors.New("invalid security identifier keyword")
)
