package configs

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Db   Dbconfig
	Auth Authconfig
}

type Dbconfig struct {
	Dsn string
}

type Authconfig struct {
	Secret string
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}

	return &Config{
		Db: Dbconfig{
			Dsn: os.Getenv("DSN"),
		},
		Auth: Authconfig{
			Secret: os.Getenv("SECRET"),
		},
	}
}
