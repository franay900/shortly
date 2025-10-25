package auth

import "url/short/pkg/errors"

// Предопределенные ошибки для модуля аутентификации
var (
	ErrUserExists       = errors.ErrUserExists
	ErrWrongCredentials = errors.ErrWrongCredentials
)
