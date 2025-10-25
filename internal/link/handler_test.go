package link

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"url/short/pkg/event"
	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"url/short/pkg/db"
)

func TestLinkHandler_Create(t *testing.T) {
	// Настройка мока
	database, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	
	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: database,
	}))
	if err != nil {
		t.Fatal(err)
	}
	
	// Создаем тестовые зависимости
	linkRepo := NewLinkRepository(&db.DB{DB: gormDB})
	eventBus := event.NewEventBus()
	
	handler := &LinkHandler{
		LinkRepository: linkRepo,
		EventBus:       eventBus,
	}
	
	// Настройка ожиданий для проверки уникальности хеша
	mock.ExpectQuery("SELECT").WillReturnError(gorm.ErrRecordNotFound)
	
	// Настройка ожиданий для создания ссылки
	mock.ExpectBegin()
	insertRows := sqlmock.NewRows([]string{"id"}).AddRow(1)
	mock.ExpectQuery(`INSERT INTO "links"`).WillReturnRows(insertRows)
	mock.ExpectCommit()
	
	// Тестовые данные
	createReq := LinkCreateRequest{
		Url: "https://example.com",
	}
	data, err := json.Marshal(createReq)
	if err != nil {
		t.Fatal(err)
	}
	
	// Создаем запрос
	req, err := http.NewRequest(http.MethodPost, "/link", bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	
	// Создаем ResponseWriter
	w := httptest.NewRecorder()
	
	// Выполняем тест
	handler.Create()(w, req)
	
	// Проверки
	if w.Code != http.StatusCreated {
		t.Fatalf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}
	
	// Проверяем, что все ожидания выполнены
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

func TestLinkHandler_GetAll(t *testing.T) {
	// Настройка мока
	database, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	
	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: database,
	}))
	if err != nil {
		t.Fatal(err)
	}
	
	// Создаем тестовые зависимости
	linkRepo := NewLinkRepository(&db.DB{DB: gormDB})
	eventBus := event.NewEventBus()
	
	handler := &LinkHandler{
		LinkRepository: linkRepo,
		EventBus:       eventBus,
	}
	
	// Настройка ожиданий для подсчета
	countRows := sqlmock.NewRows([]string{"count"}).AddRow(2)
	mock.ExpectQuery("SELECT count").WillReturnRows(countRows)
	
	// Настройка ожиданий для получения ссылок
	linkRows := sqlmock.NewRows([]string{"id", "url", "hash"}).
		AddRow(1, "https://example.com", "abc123").
		AddRow(2, "https://test.com", "def456")
	mock.ExpectQuery("SELECT").WillReturnRows(linkRows)
	
	// Создаем запрос с параметрами
	req, err := http.NewRequest(http.MethodGet, "/link?limit=10&offset=0", nil)
	if err != nil {
		t.Fatal(err)
	}
	
	// Создаем ResponseWriter
	w := httptest.NewRecorder()
	
	// Выполняем тест
	handler.GetAll()(w, req)
	
	// Проверки
	if w.Code != http.StatusOK {
		t.Fatalf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
	
	// Проверяем, что все ожидания выполнены
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

func TestLinkHandler_GetAll_InvalidParameters(t *testing.T) {
	// Настройка мока
	database, _, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	
	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: database,
	}))
	if err != nil {
		t.Fatal(err)
	}
	
	// Создаем тестовые зависимости
	linkRepo := NewLinkRepository(&db.DB{DB: gormDB})
	eventBus := event.NewEventBus()
	
	handler := &LinkHandler{
		LinkRepository: linkRepo,
		EventBus:       eventBus,
	}
	
	// Тест с невалидным limit
	req, err := http.NewRequest(http.MethodGet, "/link?limit=invalid&offset=0", nil)
	if err != nil {
		t.Fatal(err)
	}
	
	w := httptest.NewRecorder()
	handler.GetAll()(w, req)
	
	if w.Code != http.StatusBadRequest {
		t.Fatalf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	
	// Тест с отрицательным offset
	req, err = http.NewRequest(http.MethodGet, "/link?limit=10&offset=-1", nil)
	if err != nil {
		t.Fatal(err)
	}
	
	w = httptest.NewRecorder()
	handler.GetAll()(w, req)
	
	if w.Code != http.StatusBadRequest {
		t.Fatalf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	
	// Тест с превышением лимита
	req, err = http.NewRequest(http.MethodGet, "/link?limit=2000&offset=0", nil)
	if err != nil {
		t.Fatal(err)
	}
	
	w = httptest.NewRecorder()
	handler.GetAll()(w, req)
	
	if w.Code != http.StatusBadRequest {
		t.Fatalf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestLinkHandler_GoTo(t *testing.T) {
	// Настройка мока
	database, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	
	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: database,
	}))
	if err != nil {
		t.Fatal(err)
	}
	
	// Создаем тестовые зависимости
	linkRepo := NewLinkRepository(&db.DB{DB: gormDB})
	eventBus := event.NewEventBus()
	
	handler := &LinkHandler{
		LinkRepository: linkRepo,
		EventBus:       eventBus,
	}
	
	// Настройка ожиданий
	linkRows := sqlmock.NewRows([]string{"id", "url", "hash"}).
		AddRow(1, "https://example.com", "abc123")
	mock.ExpectQuery("SELECT").WillReturnRows(linkRows)
	
	// Создаем запрос
	req, err := http.NewRequest(http.MethodGet, "/abc123", nil)
	if err != nil {
		t.Fatal(err)
	}
	
	// Создаем ResponseWriter
	w := httptest.NewRecorder()
	
	// Выполняем тест
	handler.GoTo()(w, req)
	
	// Проверки
	if w.Code != http.StatusTemporaryRedirect {
		t.Fatalf("Expected status %d, got %d", http.StatusTemporaryRedirect, w.Code)
	}
	
	// Проверяем, что все ожидания выполнены
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}
