package req

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"url/short/pkg/errors"
)

func TestHandleBody_Success(t *testing.T) {
	// Тестовые данные
	testData := `{"email":"test@example.com","password":"password123","name":"Test User"}`
	
	// Создаем запрос
	req, err := http.NewRequest(http.MethodPost, "/test", bytes.NewReader([]byte(testData)))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	
	// Создаем ResponseWriter
	w := httptest.NewRecorder()
	
	// Выполняем тест
	result, err := HandleBody[TestRequest](w, req)
	
	// Проверки
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
	
	if result.Email != "test@example.com" {
		t.Fatalf("Expected email 'test@example.com', got '%s'", result.Email)
	}
	
	if result.Password != "password123" {
		t.Fatalf("Expected password 'password123', got '%s'", result.Password)
	}
	
	if result.Name != "Test User" {
		t.Fatalf("Expected name 'Test User', got '%s'", result.Name)
	}
}

func TestHandleBody_InvalidJSON(t *testing.T) {
	// Невалидный JSON
	invalidJSON := `{"email":"test@example.com","password":`
	
	// Создаем запрос
	req, err := http.NewRequest(http.MethodPost, "/test", bytes.NewReader([]byte(invalidJSON)))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	
	// Создаем ResponseWriter
	w := httptest.NewRecorder()
	
	// Выполняем тест
	result, err := HandleBody[TestRequest](w, req)
	
	// Проверки
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	
	if result != nil {
		t.Fatal("Expected nil result, got result")
	}
	
	// Проверяем статус код
	if w.Code != http.StatusBadRequest {
		t.Fatalf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	
	// Проверяем тип ошибки
	appErr, ok := err.(*errors.AppError)
	if !ok {
		t.Fatal("Expected AppError, got different error type")
	}
	
	if appErr.Code != errors.ErrCodeValidationFailed {
		t.Fatalf("Expected error code '%s', got '%s'", errors.ErrCodeValidationFailed, appErr.Code)
	}
}

func TestHandleBody_ValidationFailed(t *testing.T) {
	// Валидный JSON, но невалидные данные (отсутствует email)
	testData := `{"password":"password123","name":"Test User"}`
	
	// Создаем запрос
	req, err := http.NewRequest(http.MethodPost, "/test", bytes.NewReader([]byte(testData)))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	
	// Создаем ResponseWriter
	w := httptest.NewRecorder()
	
	// Выполняем тест
	result, err := HandleBody[TestRequest](w, req)
	
	// Проверки
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	
	if result != nil {
		t.Fatal("Expected nil result, got result")
	}
	
	// Проверяем статус код
	if w.Code != http.StatusBadRequest {
		t.Fatalf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	
	// Проверяем тип ошибки
	appErr, ok := err.(*errors.AppError)
	if !ok {
		t.Fatal("Expected AppError, got different error type")
	}
	
	if appErr.Code != errors.ErrCodeValidationFailed {
		t.Fatalf("Expected error code '%s', got '%s'", errors.ErrCodeValidationFailed, appErr.Code)
	}
	
	// Проверяем, что есть детали ошибки
	if appErr.Details == "" {
		t.Fatal("Expected error details, got empty string")
	}
}

func TestHandleBody_EmptyBody(t *testing.T) {
	// Пустое тело запроса
	req, err := http.NewRequest(http.MethodPost, "/test", bytes.NewReader([]byte("")))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	
	// Создаем ResponseWriter
	w := httptest.NewRecorder()
	
	// Выполняем тест
	result, err := HandleBody[TestRequest](w, req)
	
	// Проверки
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	
	if result != nil {
		t.Fatal("Expected nil result, got result")
	}
	
	// Проверяем статус код
	if w.Code != http.StatusBadRequest {
		t.Fatalf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// Тестовый тип для валидации
type TestRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
	Name     string `json:"name" validate:"required"`
}
