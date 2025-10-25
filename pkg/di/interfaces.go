package di

import (
	"time"
	"url/short/internal/link"
	"url/short/internal/stat"
	"url/short/internal/user"
)

// IStatRepository интерфейс для репозитория статистики
type IStatRepository interface {
	AddClick(linkId uint)
	GetStats(by string, from, to time.Time) []stat.GetStatResponse
}

// IUserRepository интерфейс для репозитория пользователей
type IUserRepository interface {
	Create(user *user.User) (*user.User, error)
	FindByEmail(email string) (*user.User, error)
}

// ILinkRepository интерфейс для репозитория ссылок
type ILinkRepository interface {
	Create(link *link.Link) (*link.Link, error)
	GetByHash(hash string) (*link.Link, error)
	Update(link *link.Link) (*link.Link, error)
	Delete(id uint) error
	GetById(id uint) (*link.Link, error)
	Count() int64
	Get(limit, offset int) []link.Link
}

// IAuthService интерфейс для сервиса аутентификации
type IAuthService interface {
	Login(email, password string) (string, error)
	Register(email, password, name string) (string, error)
}
