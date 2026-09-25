package config

import (
	"errors"
	"fmt"
	"os"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	HTTPAddr       string
	DatabaseURL    string
	DBMaxConns     int32
	RequestTimeout time.Duration
	ValkeyAddr	 string
	BlockedCIDRs []string
	PDFRoot string
	StatsCacheTTL time.Duration
}

func Load() (Config, error) {

	blockedCIDRs := []string{}
	if value := os.Getenv("BLOCKED_CIDRS"); value != "" {
		blockedCIDRs = strings.Split(value, ",")
	}

	cfg := Config{
		HTTPAddr:    getenv("HTTP_ADDR", ":8080"),
		DatabaseURL: DatabaseURLFromEnv(),
		ValkeyAddr:  getenv("VALKEY_ADDR", "localhost:6379"),
		BlockedCIDRs: blockedCIDRs,
		PDFRoot: getenv("PDF_ROOT", "./pdfs"),
	}

	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}

	n, err := strconv.ParseInt(getenv("DB_MAX_CONNS", "20"), 10, 32)
	if err != nil || n < 1 {
		return Config{}, fmt.Errorf("invalid DB_MAX_CONNS")
	}
	cfg.DBMaxConns = int32(n)

	cfg.RequestTimeout, err = time.ParseDuration(getenv("REQUEST_TIMEOUT", "2s"))
	if err != nil {
		return Config{}, fmt.Errorf("invalid REQUEST_TIMEOUT: %w", err)
	}

	cfg.StatsCacheTTL, err = time.ParseDuration(getenv("CACHE_TTL", "5m"))
	if err != nil {
		return Config{}, fmt.Errorf("invalid CACHE_TTL: %w", err)
	}

	return cfg, nil
}

func DatabaseURLFromEnv() string {
	if dsn := os.Getenv("DATABASE_URL"); dsn != "" {
		return dsn
	}
	password, ok := os.LookupEnv("POSTGRES_PASSWORD")
	if !ok || password == "" {
		return ""
	}
	user := getenv("POSTGRES_USER", "zishan044")
	database := getenv("POSTGRES_DB", "results")
	host := getenv("POSTGRES_HOST", "postgres")
	port := getenv("POSTGRES_PORT", "5432")
	u := &url.URL{Scheme: "postgres", User: url.UserPassword(user, password), Host: net.JoinHostPort(host, port), Path: database}
	query := u.Query()
	query.Set("sslmode", "disable")
	u.RawQuery = query.Encode()
	return u.String()
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}