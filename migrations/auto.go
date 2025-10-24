package main

import (
	"os"
	"url/short/internal/link"
	"url/short/internal/stat"
	"url/short/internal/user"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		panic("failed to load env")
	}

	db, err := gorm.Open(postgres.Open(os.Getenv("DSN")), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	db.AutoMigrate(&link.Link{}, &user.User{}, &stat.Stat{})

}
