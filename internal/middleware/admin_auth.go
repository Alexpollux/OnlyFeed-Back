package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ArthurDelaporte/OnlyFeed-Back/internal/logs"
	"github.com/ArthurDelaporte/OnlyFeed-Back/internal/user"
)

// AdminOnlyMiddleware permet de protéger certaines routes aux admins uniquement
func AdminOnlyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		route := c.FullPath()
		userID := c.GetString("user_id")

		if userID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Utilisateur non authentifié"})
			logs.LogJSON("WARN", "Non-authenticated user tried admin route", map[string]interface{}{
				"route": route,
			})
			return
		}

		isAdmin, err := user.IsAdmin(userID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Erreur vérification admin"})
			logs.LogJSON("ERROR", "Erreur DB admin check", map[string]interface{}{
				"error":  err.Error(),
				"route":  route,
				"userID": userID,
			})
			return
		}

		if !isAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Accès réservé aux administrateurs"})
			logs.LogJSON("WARN", "Non-admin user blocked from admin route", map[string]interface{}{
				"route":  route,
				"userID": userID,
			})
			return
		}

		c.Next()
	}
}
