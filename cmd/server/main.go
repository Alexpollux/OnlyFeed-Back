package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/ArthurDelaporte/OnlyFeed-Back/internal/auth"
	"github.com/ArthurDelaporte/OnlyFeed-Back/internal/database"
	"github.com/ArthurDelaporte/OnlyFeed-Back/internal/follow"
	"github.com/ArthurDelaporte/OnlyFeed-Back/internal/like"
	"github.com/ArthurDelaporte/OnlyFeed-Back/internal/logs"
	"github.com/ArthurDelaporte/OnlyFeed-Back/internal/message"
	"github.com/ArthurDelaporte/OnlyFeed-Back/internal/middleware"
	"github.com/ArthurDelaporte/OnlyFeed-Back/internal/post"
	"github.com/ArthurDelaporte/OnlyFeed-Back/internal/storage"
	"github.com/ArthurDelaporte/OnlyFeed-Back/internal/stripe"
	"github.com/ArthurDelaporte/OnlyFeed-Back/internal/user"
)

func main() {
	_ = godotenv.Load()

	dsn := os.Getenv("SUPABASE_DB_URL")
	domainUrl := os.Getenv("DOMAIN_URL")
	if dsn == "" {
		panic("SUPABASE_DB_URL manquant")
	}
	database.Connect(dsn)

	if err := storage.InitS3(); err != nil {
		log.Fatalf("❌ Init S3 : %v", err)
	}

	r := gin.New()

	// Middleware de logs custom pour ignorer "/"
	r.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		if !strings.HasPrefix(param.Path, "/api/") {
			return ""
		}
		return fmt.Sprintf("[GIN] %s | %3d | %13v | %15s |%-7s %#v\n",
			param.TimeStamp.Format(time.RFC3339),
			param.StatusCode,
			param.Latency,
			param.ClientIP,
			param.Method,
			param.Path,
		)
	}))

	// Middleware recovery pour éviter que l'app crash sur panic
	r.Use(gin.Recovery())

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"https://onlyfeed-9f05c.web.app/", "http://localhost:5000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Refresh-Token"},
		ExposeHeaders:    []string{"Content-Length", "X-New-Access-Token"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Routes de debug/logs (du GitHub)
	r.GET("/info", func(c *gin.Context) {
		route := c.FullPath()
		c.JSON(200, gin.H{"status": "INFO"})
		logs.LogJSON("INFO", "INFOOOOO", map[string]interface{}{
			"route": route,
			"extra": fmt.Sprintf("DOMAIN_URL is : %s", domainUrl),
		})
	})
	r.GET("/warn", func(c *gin.Context) {
		route := c.FullPath()
		c.JSON(200, gin.H{"status": "WARN"})
		logs.LogJSON("WARN", "WAAAAARN", map[string]interface{}{
			"route": route,
		})
	})
	r.GET("/error", func(c *gin.Context) {
		route := c.FullPath()
		c.JSON(200, gin.H{"status": "ERROR"})
		logs.LogJSON("ERROR", "ERROOOOR", map[string]interface{}{
			"route": route,
		})
	})
	r.GET("/fatal", func(c *gin.Context) {
		route := c.FullPath()
		c.JSON(200, gin.H{"status": "FATAL"})
		logs.LogJSON("FATAL", "FATAAAAAL", map[string]interface{}{
			"route": route,
		})
	})

	api := r.Group("/api")

	// /api/auth
	apiAuth := api.Group("/auth")
	apiAuth.POST("/signup", auth.Signup)
	apiAuth.POST("/login", auth.Login)

	// Appeler uniquement par stripe donc pas de token
	api.POST("/stripe/webhook", stripe.HandleStripeWebhook)

	api.GET("/users/search", user.SearchUsers)

	// authentification optionnelle
	api.Use(middleware.OptionalAuthMiddleware())

	// /api/users/username
	apiUsersUsername := api.Group("/users/username")
	apiUsersUsername.GET("/:username", user.GetUserByUsername)
	apiUsersUsername.GET("/:username/posts", post.GetPostsByUsername)

	// Routes publiques pour les posts
	api.GET("/posts", like.GetPostsWithLikes)
	api.GET("/posts/:id/comments", post.GetCommentsByPostID)
	api.GET("/posts/:id/likes", like.GetLikeStatus)
	api.GET("/posts/:id", like.GetPostByIDWithLikes)

	// Routes protégées par authentification
	api.Use(middleware.AuthMiddleware())

	api.POST("/auth/logout", auth.Logout)

	// /api/me
	apiMe := api.Group("/me")
	apiMe.GET("", user.GetMe)
	apiMe.PUT("", user.UpdateMe)

	// /api/users
	apiUsers := api.Group("/users")
	apiUsers.GET("/:id", user.GetUser)
	apiUsers.PUT("/:id", user.UpdateUser)
	apiUsers.DELETE("/:id", user.DeleteUser)

	// Routes pour les posts nécessitant une authentification
	apiPosts := api.Group("/posts")
	apiPosts.POST("", post.CreatePost)
	apiPosts.GET("/me", post.GetUserPosts)
	apiPosts.DELETE("/:id", post.DeletePost)
	apiPosts.POST("/:id/like", like.ToggleLike)

	// Routes pour les commentaires nécessitant une authentification
	apiComments := api.Group("/comments")
	apiComments.POST("", post.CreateComment)
	apiComments.DELETE("/:id", post.DeleteComment)

	// Routes pour la messagerie
	apiMessages := api.Group("/messages")
	apiMessages.GET("/conversations", message.GetConversations)
	apiMessages.GET("/conversations/:id", message.GetConversationMessages)
	apiMessages.POST("/send", message.SendMessage)
	apiMessages.PUT("/:id/read", message.MarkMessageAsRead)
	apiMessages.DELETE("/:id", message.DeleteMessage)
	apiMessages.DELETE("/conversations/:id", message.DeleteConversation)

	// /api/follow
	apiFollow := api.Group("/follow")
	apiFollow.POST("/:id", follow.FollowUser)
	apiFollow.DELETE("/:id", follow.UnfollowUser)
	apiFollow.GET("/", follow.GetFollowing)
	apiFollow.GET("/followers/:id", follow.GetFollowers)

	stripeGroup := api.Group("/stripe")
	stripeGroup.POST("/create-account-link", stripe.CreateAccountLink)
	stripeGroup.GET("/complete-connect", stripe.CompleteConnect)
	stripeGroup.POST("/create-subscription-session/:creator_id", stripe.CreateSubscriptionSession)
	stripeGroup.DELETE("/unsubscribe/:creator_id", stripe.Unsubscribe)

	apiAdmin := api.Group("/admin")
	apiAdmin.Use(middleware.AdminOnlyMiddleware())

	err := r.Run("0.0.0.0:8080")
	if err != nil {
		return
	}
}
