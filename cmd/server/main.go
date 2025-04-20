package main

import (
	"os"

	"github.com/ArthurDelaporte/OnlyFeed-Back/internal/auth"
	"github.com/ArthurDelaporte/OnlyFeed-Back/internal/database"
	"github.com/ArthurDelaporte/OnlyFeed-Back/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	dsn := os.Getenv("SUPABASE_DB_URL")
	if dsn == "" {
		panic("SUPABASE_DB_URL manquant")
	}

	database.Connect(dsn)

	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	api := r.Group("/api")

	// Inscription & Connexion
	api.POST("/signup", auth.Signup)
	api.POST("/login", auth.Login)

	api.Use(middleware.AuthMiddleware())
	api.GET("/me", func(c *gin.Context) {
		userID := c.GetString("user_id")
		c.JSON(200, gin.H{"user_id": userID})
	})

	err := r.Run(":8080")
	if err != nil {
		return
	}
}
