package stat

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"url/short/pkg/db"
)

func TestStatHandler_GetStat(t *testing.T) {
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
	statRepo := NewStatRepository(&db.DB{DB: gormDB})
	
	handler := &StatHandler{
		StatRepository: statRepo,
	}
	
	// Настройка ожиданий
	statRows := sqlmock.NewRows([]string{"period", "sum"}).
		AddRow("2024-01-01", 10).
		AddRow("2024-01-02", 15)
	mock.ExpectQuery("SELECT").WillReturnRows(statRows)
	
	// Создаем запрос
	req, err := http.NewRequest(http.MethodGet, "/stat?from=2024-01-01&to=2024-01-02&by=day", nil)
	if err != nil {
		t.Fatal(err)
	}
	
	// Создаем ResponseWriter
	w := httptest.NewRecorder()
	
	// Выполняем тест
	handler.GetStat()(w, req)
	
	// Проверки
	if w.Code != http.StatusOK {
		t.Fatalf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
	
	// Проверяем, что все ожидания выполнены
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

func TestStatHandler_GetStat_InvalidDate(t *testing.T) {
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
	statRepo := NewStatRepository(&db.DB{DB: gormDB})
	
	handler := &StatHandler{
		StatRepository: statRepo,
	}
	
	// Тест с невалидной датой from
	req, err := http.NewRequest(http.MethodGet, "/stat?from=invalid-date&to=2024-01-02&by=day", nil)
	if err != nil {
		t.Fatal(err)
	}
	
	w := httptest.NewRecorder()
	handler.GetStat()(w, req)
	
	if w.Code != http.StatusBadRequest {
		t.Fatalf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	
	// Тест с невалидной датой to
	req, err = http.NewRequest(http.MethodGet, "/stat?from=2024-01-01&to=invalid-date&by=day", nil)
	if err != nil {
		t.Fatal(err)
	}
	
	w = httptest.NewRecorder()
	handler.GetStat()(w, req)
	
	if w.Code != http.StatusBadRequest {
		t.Fatalf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestStatHandler_GetStat_InvalidGroupBy(t *testing.T) {
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
	statRepo := NewStatRepository(&db.DB{DB: gormDB})
	
	handler := &StatHandler{
		StatRepository: statRepo,
	}
	
	// Тест с невалидным параметром by
	req, err := http.NewRequest(http.MethodGet, "/stat?from=2024-01-01&to=2024-01-02&by=invalid", nil)
	if err != nil {
		t.Fatal(err)
	}
	
	w := httptest.NewRecorder()
	handler.GetStat()(w, req)
	
	if w.Code != http.StatusBadRequest {
		t.Fatalf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestStatHandler_GetStat_MonthGroupBy(t *testing.T) {
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
	statRepo := NewStatRepository(&db.DB{DB: gormDB})
	
	handler := &StatHandler{
		StatRepository: statRepo,
	}
	
	// Настройка ожиданий для группировки по месяцам
	statRows := sqlmock.NewRows([]string{"period", "sum"}).
		AddRow("2024-01", 25).
		AddRow("2024-02", 30)
	mock.ExpectQuery("SELECT").WillReturnRows(statRows)
	
	// Создаем запрос
	req, err := http.NewRequest(http.MethodGet, "/stat?from=2024-01-01&to=2024-02-28&by=month", nil)
	if err != nil {
		t.Fatal(err)
	}
	
	// Создаем ResponseWriter
	w := httptest.NewRecorder()
	
	// Выполняем тест
	handler.GetStat()(w, req)
	
	// Проверки
	if w.Code != http.StatusOK {
		t.Fatalf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
	
	// Проверяем, что все ожидания выполнены
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}
