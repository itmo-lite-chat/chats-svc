package config

import "os"

type Config struct {
	GRPCAddr    string
	PostgresDSN string
}

func New() Config {
	return Config{
		GRPCAddr:    getenv("GRPC_ADDR", ":9992"),
		PostgresDSN: getenv("POSTGRES_DSN", "postgres://postgres:postgres@127.0.0.1:5433/postgres?sslmode=disable"),
	}
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
