package main

import (
	"bytes"
	"encoding/json"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"url/short/internal/auth"
	"url/short/internal/user"
)

func initDb() *gorm.DB {
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Error loading .env file")
	}
	db, err := gorm.Open(postgres.Open(os.Getenv("DSN")), &gorm.Config{})

	if err != nil {
		panic("failed to connect database")
	}
	return db
}

func initData(db *gorm.DB) {
	db.Create(&user.User{
		Email:    "email4@mail.ru",
		Password: "$2a$10$xwLLgG77tJ5x9hWAXJrk0OFq/bpY4i9pojqsmxLyznn45A5.COVb6",
		Name:     "user",
	})
}

func deleteData(db *gorm.DB) {
	db.Unscoped().
		Where("email = ?", "email4@mail.ru").
		Delete(&user.User{})
}

func TestLoginSuccess(t *testing.T) {
	// prepare
	db := initDb()
	initData(db)

	ts := httptest.NewServer(App())
	defer ts.Close()
	data, _ := json.Marshal(&auth.LoginRequest{
		Email:    "email4@mail.ru",
		Password: "123",
	})

	res, err := http.Post(ts.URL+"/auth/login", "application/json", bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != http.StatusOK {
		t.Fatalf("got %d, want %d", res.StatusCode, http.StatusOK)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	var resData auth.LoginResponse

	if err := json.Unmarshal(body, &resData); err != nil {
		t.Fatal(err)
	}

	if resData.Token == "" {
		t.Fatal("got no token")
	}

	deleteData(db)

}

func TestLoginFailed(t *testing.T) {
	// prepare
	db := initDb()
	initData(db)

	ts := httptest.NewServer(App())
	defer ts.Close()
	data, _ := json.Marshal(&auth.LoginRequest{
		Email:    "email3@mail.ru",
		Password: "2",
	})

	res, err := http.Post(ts.URL+"/auth/login", "application/json", bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("got %d, want %d", res.StatusCode, http.StatusUnauthorized)
	}

	deleteData(db)
}
