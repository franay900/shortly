package auth

import (
	"url/short/internal/user"
	"url/short/pkg/di"
	"url/short/pkg/errors"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	UserRepository di.IUserRepository
}

func NewAuthService(userRepository di.IUserRepository) *AuthService {
	return &AuthService{UserRepository: userRepository}
}

func (service *AuthService) Login(email, password string) (string, error) {
	existedUser, err := service.UserRepository.FindByEmail(email)
	if err != nil {
		return "", errors.WrapError(err, errors.ErrCodeDatabaseError, "Failed to find user")
	}

	if existedUser == nil {
		return "", ErrWrongCredentials
	}

	err = bcrypt.CompareHashAndPassword([]byte(existedUser.Password), []byte(password))
	if err != nil {
		return "", ErrWrongCredentials
	}

	return existedUser.Email, nil
}

func (service *AuthService) Register(email, password, name string) (string, error) {
	existedUser, err := service.UserRepository.FindByEmail(email)
	if err != nil {
		return "", errors.WrapError(err, errors.ErrCodeDatabaseError, "Failed to check if user exists")
	}

	if existedUser != nil {
		return "", ErrUserExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", errors.WrapError(err, errors.ErrCodeInternalError, "Failed to hash password")
	}

	user := &user.User{
		Email:    email,
		Password: string(hashedPassword),
		Name:     name,
	}

	_, err = service.UserRepository.Create(user)
	if err != nil {
		return "", errors.WrapError(err, errors.ErrCodeDatabaseError, "Failed to create user")
	}

	return user.Email, nil
}
