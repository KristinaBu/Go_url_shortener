package config

import (
	"flag"
	"fmt"
	"os"
)

const (
	StorageMemory   = "memory"
	StoragePostgres = "postgres"
)

type Config struct {
	HTTPAddr  string
	Storage   string
	LogOutput string
	Database  string
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

	return Config{
		HTTPAddr:  *httpAddr,
		Storage:   *storage,
		LogOutput: *logOutput,
		Database:  *database,
	}, nil
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}

func OpenLogOutput(path string) (*os.File, error) {
	if path == "stdout" {
		return os.Stdout, nil
	}

	return os.OpenFile(
		path,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0644,
	)
}
