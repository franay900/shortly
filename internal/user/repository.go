package user

import "url/short/pkg/db"

type UserRepository struct {
	database *db.DB
}

func NewUserRepository(database *db.DB) *UserRepository {
	return &UserRepository{database: database}
}

func (repo *UserRepository) Create(user *User) (*User, error) {
	result := repo.database.DB.Create(user)
	if result.Error != nil {
		return nil, result.Error
	}

	return user, nil
}

func (repo *UserRepository) FindByEmail(email string) (*User, error) {
	var user User
	result := repo.database.DB.First(&user, "email = ?", email)

	if result.Error != nil {
		return nil, result.Error
	}

	return &user, nil
}
