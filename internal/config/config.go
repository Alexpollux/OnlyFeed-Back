package config

import (
	"os"
)

type Config struct {
	DBUrl     string
	JWTSecret string
	Supabase  string
}

func LoadConfig() *Config {
	return &Config{
		DBUrl:     os.Getenv("SUPABASE_DB_URL"),
		JWTSecret: os.Getenv("JWT_SECRET"),
		Supabase:  os.Getenv("NEXT_PUBLIC_SUPABASE_URL"),
	}
}
