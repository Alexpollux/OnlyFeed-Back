package database

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/ArthurDelaporte/OnlyFeed-Back/internal/logs"
)

var DB *gorm.DB

func Connect(dsn string) {
	var err error
	DB, err = gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true,
	}))
	if err != nil {
		logs.LogJSON("FATAL", "Supabase connection error", map[string]interface{}{})
	}

	logs.LogJSON("DEBUG", "Database connection established", map[string]interface{}{})
}
