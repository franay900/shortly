package auth

import (
	"bytes"
	"encoding/json"
	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	"testing"
	"url/short/configs"
	"url/short/internal/user"
	"url/short/pkg/db"
)

func bootstrap() (*AuthHandler, sqlmock.Sqlmock, error) {
	database, mock, err := sqlmock.New()

	if err != nil {
		return nil, nil, err

	}
	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: database,
	}))

	if err != nil {
		return nil, nil, err

	}

	userRepository := user.NewUserRepository(&db.DB{
		DB: gormDB,
	})

	handler := AuthHandler{
		Config: &configs.Config{
			Auth: configs.Authconfig{
				Secret: "secret",
			},
		},
		AuthService: &AuthService{
			UserRepository: userRepository,
		},
	}

	return &handler, mock, nil
}

func TestLoginHandlerSuccess(t *testing.T) {
	handler, mock, err := bootstrap()
	if err != nil {
		t.Fatal(err)
	}

	rows := sqlmock.NewRows([]string{"email", "password"}).AddRow("email4@mail.ru", "$2a$10$xwLLgG77tJ5x9hWAXJrk0OFq/bpY4i9pojqsmxLyznn45A5.COVb6")
	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	data, err := json.Marshal(&LoginRequest{
		Email:    "email4@mail.ru",
		Password: "123",
	})
	if err != nil {
		t.Fatal(err)
	}

	reader := bytes.NewReader(data)
	w := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodPost, "/auth/login", reader)
	if err != nil {
		t.Fatal(err)
	}

	handler.Login()(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestRegisterHandlerSuccess(t *testing.T) {
	handler, mock, err := bootstrap()
	if err != nil {
		t.Fatal(err)
	}

	rows := sqlmock.NewRows([]string{"email", "password"})
	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	mock.ExpectBegin()
	insertRows := sqlmock.NewRows([]string{"id"}).AddRow(1)
	mock.ExpectQuery(`INSERT INTO "users"`).WillReturnRows(insertRows)
	mock.ExpectCommit()

	data, err := json.Marshal(&RegisterRequest{
		Email:    "email4@mail.ru",
		Password: "123",
		Name:     "user",
	})
	if err != nil {
		t.Fatal(err)
	}

	reader := bytes.NewReader(data)
	w := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodPost, "/auth/register", reader)
	if err != nil {
		t.Fatal(err)
	}

	handler.Register()(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}
}
