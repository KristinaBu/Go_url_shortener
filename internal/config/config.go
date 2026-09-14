package config

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"
)

const (
	StorageMemory    = "memory"
	StoragePostgres  = "postgres"
	DefaultCacheSize = 1000
)

type Config struct {
	HTTPAddr          string
	Storage           string
	LogOutput         string
	Database          string
	CacheSize         int
	ReadHeaderTimeout time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
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

	cacheSizeDefault, err := getEnvInt("CACHE_SIZE", DefaultCacheSize)
	if err != nil {
		return Config{}, err
	}

	cacheSize := flag.Int(
		"cache-size",
		cacheSizeDefault,
		"LRU cache capacity",
	)

	readHeaderTimeout := flag.Duration(
		"read-header-timeout",
		5*time.Second,
		"maximum time to read request headers",
	)

	readTimeout := flag.Duration(
		"read-timeout",
		10*time.Second,
		"maximum duration for reading the request",
	)

	writeTimeout := flag.Duration(
		"write-timeout",
		10*time.Second,
		"maximum duration before timing out writes",
	)

	idleTimeout := flag.Duration(
		"idle-timeout",
		60*time.Second,
		"maximum amount of time to wait for the next request",
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
		HTTPAddr:          *httpAddr,
		Storage:           *storage,
		LogOutput:         *logOutput,
		Database:          *database,
		CacheSize:         *cacheSize,
		ReadHeaderTimeout: *readHeaderTimeout,
		ReadTimeout:       *readTimeout,
		WriteTimeout:      *writeTimeout,
		IdleTimeout:       *idleTimeout,
	}, nil
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}

func getEnvInt(key string, fallback int) (int, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}

	result, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf(
			"%s must be an integer: %w",
			key,
			err,
		)
	}

	return result, nil
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
