package errors

import (
	"encoding/json"
	"net/http"
)

// AppError представляет структурированную ошибку приложения
type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// Error реализует интерфейс error
func (e *AppError) Error() string {
	return e.Message
}

// NewAppError создает новую ошибку приложения
func NewAppError(code, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

// WithDetails добавляет детали к ошибке
func (e *AppError) WithDetails(details string) *AppError {
	e.Details = details
	return e
}

// Предопределенные коды ошибок
const (
	// Аутентификация
	ErrCodeUserExists        = "USER_EXISTS"
	ErrCodeWrongCredentials  = "WRONG_CREDENTIALS"
	ErrCodeUnauthorized      = "UNAUTHORIZED"
	ErrCodeInvalidToken      = "INVALID_TOKEN"
	
	// Валидация
	ErrCodeValidationFailed  = "VALIDATION_FAILED"
	ErrCodeInvalidParameter  = "INVALID_PARAMETER"
	ErrCodeMissingParameter  = "MISSING_PARAMETER"
	
	// Ресурсы
	ErrCodeNotFound          = "NOT_FOUND"
	ErrCodeAlreadyExists     = "ALREADY_EXISTS"
	ErrCodeConflict          = "CONFLICT"
	
	// Сервер
	ErrCodeInternalError     = "INTERNAL_ERROR"
	ErrCodeDatabaseError     = "DATABASE_ERROR"
	ErrCodeExternalService   = "EXTERNAL_SERVICE_ERROR"
)

// Предопределенные ошибки
var (
	ErrUserExists       = NewAppError(ErrCodeUserExists, "User already exists")
	ErrWrongCredentials = NewAppError(ErrCodeWrongCredentials, "Invalid email or password")
	ErrUnauthorized     = NewAppError(ErrCodeUnauthorized, "Unauthorized access")
	ErrInvalidToken     = NewAppError(ErrCodeInvalidToken, "Invalid or expired token")
	ErrNotFound         = NewAppError(ErrCodeNotFound, "Resource not found")
	ErrValidationFailed = NewAppError(ErrCodeValidationFailed, "Validation failed")
	ErrInternalError    = NewAppError(ErrCodeInternalError, "Internal server error")
)

// WriteError записывает ошибку в HTTP ответ
func WriteError(w http.ResponseWriter, err error, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	var appErr *AppError
	if e, ok := err.(*AppError); ok {
		appErr = e
	} else {
		// Если это не AppError, создаем общую ошибку
		appErr = NewAppError(ErrCodeInternalError, "Internal server error")
	}
	
	json.NewEncoder(w).Encode(appErr)
}

// GetStatusCode возвращает HTTP статус код для ошибки
func GetStatusCode(err error) int {
	if appErr, ok := err.(*AppError); ok {
		switch appErr.Code {
		case ErrCodeUserExists, ErrCodeAlreadyExists:
			return http.StatusConflict
		case ErrCodeWrongCredentials, ErrCodeUnauthorized, ErrCodeInvalidToken:
			return http.StatusUnauthorized
		case ErrCodeValidationFailed, ErrCodeInvalidParameter, ErrCodeMissingParameter:
			return http.StatusBadRequest
		case ErrCodeNotFound:
			return http.StatusNotFound
		case ErrCodeConflict:
			return http.StatusConflict
		case ErrCodeInternalError, ErrCodeDatabaseError, ErrCodeExternalService:
			return http.StatusInternalServerError
		default:
			return http.StatusInternalServerError
		}
	}
	return http.StatusInternalServerError
}

// WrapError оборачивает обычную ошибку в AppError
func WrapError(err error, code, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Details: err.Error(),
	}
}
