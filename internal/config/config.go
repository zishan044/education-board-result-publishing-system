package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	HTTPAddr       string
	DatabaseURL    string
	DBMaxConns     int32
}

func Load() (Config, error) {
	cfg := Config{
		HTTPAddr:    getenv("HTTP_ADDR", ":8080"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
	}

	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}

	n, err := strconv.ParseInt(getenv("DB_MAX_CONNS", "20"), 10, 32)
	if err != nil || n < 1 {
		return Config{}, fmt.Errorf("invalid DB_MAX_CONNS")
	}
	cfg.DBMaxConns = int32(n)
	
	return cfg, nil
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}