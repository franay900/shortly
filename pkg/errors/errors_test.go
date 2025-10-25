package errors

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAppError_Error(t *testing.T) {
	err := NewAppError("TEST_ERROR", "Test error message")
	
	if err.Error() != "Test error message" {
		t.Fatalf("Expected 'Test error message', got '%s'", err.Error())
	}
	
	if err.Code != "TEST_ERROR" {
		t.Fatalf("Expected 'TEST_ERROR', got '%s'", err.Code)
	}
}

func TestAppError_WithDetails(t *testing.T) {
	err := NewAppError("TEST_ERROR", "Test error message")
	err = err.WithDetails("Additional details")
	
	if err.Details != "Additional details" {
		t.Fatalf("Expected 'Additional details', got '%s'", err.Details)
	}
}

func TestGetStatusCode(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected int
	}{
		{
			name:     "User exists error",
			err:      ErrUserExists,
			expected: http.StatusConflict,
		},
		{
			name:     "Wrong credentials error",
			err:      ErrWrongCredentials,
			expected: http.StatusUnauthorized,
		},
		{
			name:     "Unauthorized error",
			err:      ErrUnauthorized,
			expected: http.StatusUnauthorized,
		},
		{
			name:     "Validation failed error",
			err:      ErrValidationFailed,
			expected: http.StatusBadRequest,
		},
		{
			name:     "Not found error",
			err:      ErrNotFound,
			expected: http.StatusNotFound,
		},
		{
			name:     "Internal error",
			err:      ErrInternalError,
			expected: http.StatusInternalServerError,
		},
		{
			name:     "Non-AppError",
			err:      &customError{message: "custom error"},
			expected: http.StatusInternalServerError,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetStatusCode(tt.err)
			if result != tt.expected {
				t.Fatalf("Expected status %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestWriteError(t *testing.T) {
	// Создаем тестовую ошибку
	err := NewAppError("TEST_ERROR", "Test error message")
	
	// Создаем ResponseWriter
	w := httptest.NewRecorder()
	
	// Выполняем тест
	WriteError(w, err, http.StatusBadRequest)
	
	// Проверки
	if w.Code != http.StatusBadRequest {
		t.Fatalf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	
	// Проверяем Content-Type
	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Fatalf("Expected Content-Type 'application/json', got '%s'", contentType)
	}
	
	// Проверяем тело ответа - проверяем что содержит нужные поля
	body := w.Body.String()
	if !contains(body, "TEST_ERROR") {
		t.Fatalf("Expected body to contain 'TEST_ERROR', got '%s'", body)
	}
	if !contains(body, "Test error message") {
		t.Fatalf("Expected body to contain 'Test error message', got '%s'", body)
	}
}

func TestWriteError_NonAppError(t *testing.T) {
	// Создаем обычную ошибку
	err := &customError{message: "custom error"}
	
	// Создаем ResponseWriter
	w := httptest.NewRecorder()
	
	// Выполняем тест
	WriteError(w, err, http.StatusInternalServerError)
	
	// Проверки
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
	
	// Проверяем тело ответа - проверяем что содержит нужные поля
	body := w.Body.String()
	if !contains(body, "INTERNAL_ERROR") {
		t.Fatalf("Expected body to contain 'INTERNAL_ERROR', got '%s'", body)
	}
	if !contains(body, "Internal server error") {
		t.Fatalf("Expected body to contain 'Internal server error', got '%s'", body)
	}
}

func TestWrapError(t *testing.T) {
	originalErr := &customError{message: "original error"}
	wrappedErr := WrapError(originalErr, "WRAPPED_ERROR", "Wrapped error message")
	
	if wrappedErr.Code != "WRAPPED_ERROR" {
		t.Fatalf("Expected code 'WRAPPED_ERROR', got '%s'", wrappedErr.Code)
	}
	
	if wrappedErr.Message != "Wrapped error message" {
		t.Fatalf("Expected message 'Wrapped error message', got '%s'", wrappedErr.Message)
	}
	
	if wrappedErr.Details != "original error" {
		t.Fatalf("Expected details 'original error', got '%s'", wrappedErr.Details)
	}
}

// Вспомогательная функция для проверки содержимого строки
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Вспомогательный тип для тестирования
type customError struct {
	message string
}

func (e *customError) Error() string {
	return e.message
}