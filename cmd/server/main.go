package main

import (
	"github.com/gin-gonic/gin"
	"log"
)

func main() {
	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "OK"})
	})

	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Erreur de lancement du serveur : %v", err)
	}
}
