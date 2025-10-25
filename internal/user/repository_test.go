package user

import (
	"testing"
	"url/short/pkg/db"
	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"errors"
)

func TestUserRepository_Create(t *testing.T) {
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
	
	repo := NewUserRepository(&db.DB{DB: gormDB})
	
	// Настройка ожиданий
	mock.ExpectBegin()
	insertRows := sqlmock.NewRows([]string{"id"}).AddRow(1)
	mock.ExpectQuery(`INSERT INTO "users"`).WillReturnRows(insertRows)
	mock.ExpectCommit()
	
	// Тестовые данные
	testUser := &User{
		Email:    "test@example.com",
		Password: "hashed_password",
		Name:     "Test User",
	}
	
	// Выполнение теста
	result, err := repo.Create(testUser)
	
	// Проверки
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if result == nil {
		t.Fatal("Expected user, got nil")
	}
	
	if result.Email != testUser.Email {
		t.Fatalf("Expected email %s, got %s", testUser.Email, result.Email)
	}
	
	// Проверяем, что все ожидания выполнены
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

func TestUserRepository_FindByEmail_Success(t *testing.T) {
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
	
	repo := NewUserRepository(&db.DB{DB: gormDB})
	
	// Настройка ожиданий
	rows := sqlmock.NewRows([]string{"id", "email", "password", "name"}).
		AddRow(1, "test@example.com", "hashed_password", "Test User")
	mock.ExpectQuery("SELECT").WillReturnRows(rows)
	
	// Выполнение теста
	result, err := repo.FindByEmail("test@example.com")
	
	// Проверки
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if result == nil {
		t.Fatal("Expected user, got nil")
	}
	
	if result.Email != "test@example.com" {
		t.Fatalf("Expected email test@example.com, got %s", result.Email)
	}
	
	// Проверяем, что все ожидания выполнены
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

func TestUserRepository_FindByEmail_NotFound(t *testing.T) {
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
	
	repo := NewUserRepository(&db.DB{DB: gormDB})
	
	// Настройка ожиданий - возвращаем ошибку "record not found"
	mock.ExpectQuery("SELECT").WillReturnError(gorm.ErrRecordNotFound)
	
	// Выполнение теста
	result, err := repo.FindByEmail("nonexistent@example.com")
	
	// Проверки
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if result != nil {
		t.Fatal("Expected nil user, got user")
	}
	
	// Проверяем, что все ожидания выполнены
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

func TestUserRepository_FindByEmail_DatabaseError(t *testing.T) {
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
	
	repo := NewUserRepository(&db.DB{DB: gormDB})
	
	// Настройка ожиданий - возвращаем ошибку БД
	dbError := errors.New("database connection failed")
	mock.ExpectQuery("SELECT").WillReturnError(dbError)
	
	// Выполнение теста
	result, err := repo.FindByEmail("test@example.com")
	
	// Проверки
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	
	if result != nil {
		t.Fatal("Expected nil user, got user")
	}
	
	if err.Error() != dbError.Error() {
		t.Fatalf("Expected error %v, got %v", dbError, err)
	}
	
	// Проверяем, что все ожидания выполнены
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}
