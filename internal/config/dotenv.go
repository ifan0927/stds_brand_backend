package config

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
)

func LoadDotenv() error {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}
