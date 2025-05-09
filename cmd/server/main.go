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
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Refresh-Token"},
		ExposeHeaders:    []string{"Content-Length", "X-New-Access-Token"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	api := r.Group("/api")

	// /api/auth
	apiAuth := api.Group("/auth")
	apiAuth.POST("/signup", auth.Signup)
	apiAuth.POST("/login", auth.Login)
	apiAuth.POST("/logout", auth.Logout)

	// /api/users/username
	apiUsersUsername := api.Group("/users/username")
	apiUsersUsername.GET("/:username", user.GetUserByUsername)

	// access_token requis
	api.Use(middleware.AuthMiddleware())

	// /api/me
	apiMe := api.Group("/me")
	apiMe.GET("", user.GetMe)
	apiMe.PUT("", user.UpdateMe)

	// /api/users
	apiUsers := api.Group("/users")
	apiUsers.GET("/:id", user.GetUser)
	apiUsers.PUT("/:id", user.UpdateUser)
	apiUsers.DELETE("/:id", user.DeleteUser)

	err := r.Run(":8080")
	if err != nil {
		return
	}
}
