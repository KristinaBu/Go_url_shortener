package config

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
)

const (
	StorageMemory    = "memory"
	StoragePostgres  = "postgres"
	defaultCacheSize = 1000
)

type Config struct {
	HTTPAddr  string
	Storage   string
	LogOutput string
	Database  string
	CacheSize int
}

type nopWriteCloser struct {
	io.Writer
}

func (nopWriteCloser) Close() error {
	return nil
}

func Parse() (Config, error) {
	httpAddr := flag.String(
		"http-addr",
		getEnv("HTTP_ADDR", ":8080"),
		"HTTP server address",
	)

	storage := flag.String(
		"storage",
		getEnv("STORAGE", StorageMemory),
		"storage implementation: memory or postgres",
	)

	logOutput := flag.String(
		"log-output",
		getEnv("LOG_OUTPUT", "stdout"),
		"log output: stdout or file path",
	)

	database := flag.String(
		"database",
		getEnv("DATABASE_URL", ""),
		"PostgreSQL connection",
	)

	cacheSize := flag.Int(
		"cache-size",
		getEnvInt("CACHE_SIZE", defaultCacheSize),
		"LRU cache capacity",
	)

	flag.Parse()

	if *storage != StorageMemory && *storage != StoragePostgres {
		return Config{}, fmt.Errorf(
			"unsupported storage %q",
			*storage,
		)
	}

	if *storage == StoragePostgres && *database == "" {
		return Config{}, fmt.Errorf(
			"database is required when storage is postgres",
		)
	}

	if *cacheSize <= 0 {
		return Config{}, fmt.Errorf(
			"cache size must be greater than zero",
		)
	}

	return Config{
		HTTPAddr:  *httpAddr,
		Storage:   *storage,
		LogOutput: *logOutput,
		Database:  *database,
		CacheSize: *cacheSize,
	}, nil
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}

func getEnvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	result, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return result
}

func OpenLogOutput(path string) (io.WriteCloser, error) {
	if path == "stdout" {
		return nopWriteCloser{Writer: os.Stdout}, nil
	}

	return os.OpenFile(
		path,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0644,
	)
}
