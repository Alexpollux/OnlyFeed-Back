package main

import (
	"github.com/ArthurDelaporte/OnlyFeed-Back/internal/config"
	"github.com/ArthurDelaporte/OnlyFeed-Back/internal/database"
	"github.com/gin-gonic/gin"
	"log"
)

func main() {
	cfg := config.LoadConfig()
	database.InitDB(cfg)

	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "OK"})
	})

	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Erreur de lancement du serveur : %v", err)
	}
}
