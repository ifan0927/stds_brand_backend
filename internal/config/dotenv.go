package config

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
)

// LoadDotenv loads local environment variables when a dotenv file is present.
func LoadDotenv() error {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}
