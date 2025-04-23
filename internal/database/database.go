package database

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
)

var DB *gorm.DB

func Connect(dsn string) {
	var err error
	DB, err = gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), // 👀 Log niveau info
	})
	if err != nil {
		log.Fatalf("Erreur de connexion à Supabase: %v", err)
	}

	//err = DB.AutoMigrate(&user.User{})
	//if err != nil {
	//	log.Fatalf("Erreur migration : %v", err)
	//}
}
