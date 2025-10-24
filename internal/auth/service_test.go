package auth

import (
	"testing"
	"url/short/internal/user"
)

type MockUserRepository struct {
}

func (m *MockUserRepository) Create(u *user.User) (*user.User, error) {
	return &user.User{
		Email: "a@mail.ru",
	}, nil
}

func (m *MockUserRepository) FindByEmail(email string) (*user.User, error) {
	return nil, nil
}

func TestRegisterSuccess(t *testing.T) {
	const initialEmail = "a@mail.ru"
	authService := NewAuthService(&MockUserRepository{})
	email, err := authService.Register(initialEmail, "1111", "Вася")

	if err != nil {
		t.Fatal(err)
	}

	if email != initialEmail {
		t.Fatalf("Expected email %s, got %s", initialEmail, email)
	}

}
