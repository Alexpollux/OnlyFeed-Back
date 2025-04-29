package main

import (
	"log"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/ArthurDelaporte/OnlyFeed-Back/internal/auth"
	"github.com/ArthurDelaporte/OnlyFeed-Back/internal/database"
	"github.com/ArthurDelaporte/OnlyFeed-Back/internal/middleware"
	"github.com/ArthurDelaporte/OnlyFeed-Back/internal/storage"
	"github.com/ArthurDelaporte/OnlyFeed-Back/internal/user"
)

func main() {
	_ = godotenv.Load()

	dsn := os.Getenv("SUPABASE_DB_URL")
	if dsn == "" {
		panic("SUPABASE_DB_URL manquant")
	}
	database.Connect(dsn)

	if err := storage.InitS3(); err != nil {
		log.Fatalf("❌ Init S3 : %v", err)
	}

	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	api := r.Group("/api")

	apiAuth := api.Group("/auth")

	// Inscription & Connexion
	apiAuth.POST("/signup", auth.Signup)
	apiAuth.POST("/login", auth.Login)

	// Oblige d'avoir un access_token en Authorization Bearer
	api.Use(middleware.AuthMiddleware())

	api.GET("/me", func(c *gin.Context) {
		userID := c.GetString("user_id")
		c.JSON(200, gin.H{"user_id": userID})
	})

	// Gestion utilisateurs
	api.GET("/users/:id", user.GetUser)
	api.PUT("/users/:id", user.UpdateUser)
	api.DELETE("/users/:id", user.DeleteUser)

	err := r.Run(":8080")
	if err != nil {
		return
	}
}
