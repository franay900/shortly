package link

import (
	"crypto/rand"
	"math/big"
	"url/short/internal/stat"

	"gorm.io/gorm"
)

type Link struct {
	gorm.Model
	Url   string      `json:"url"`
	Hash  string      `json:"hash" gorm:"uniqueIndex"`
	Stats []stat.Stat `json:"stats" gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}

func NewLink(url string) *Link {
	link := &Link{
		Url: url,
	}
	link.generateHash()
	return link
}

func (link *Link) generateHash() {
	link.Hash = RandStringRunes(6)
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

// RandStringRunes генерирует криптографически стойкую случайную строку
func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		// Используем криптографически стойкий генератор случайных чисел
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letterRunes))))
		if err != nil {
			// В случае ошибки используем fallback (не должно происходить в нормальных условиях)
			panic("failed to generate random number: " + err.Error())
		}
		b[i] = letterRunes[num.Int64()]
	}
	return string(b)
}
