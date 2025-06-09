package middleware

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func OptionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		refreshToken := c.GetHeader("X-Refresh-Token")

		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.Next()
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		jwtSecret := []byte(os.Getenv("JWT_SECRET"))

		token, _, err := new(jwt.Parser).ParseUnverified(tokenStr, jwt.MapClaims{})
		if err != nil {
			c.Next()
			return
		}

		claims := token.Claims.(jwt.MapClaims)
		expFloat, ok := claims["exp"].(float64)
		if !ok {
			c.Next()
			return
		}

		exp := int64(expFloat)
		now := time.Now().Unix()

		// Rafraîchissement si expiré
		if now > exp && refreshToken != "" {
			if newToken, err := refreshAccessToken(refreshToken); err == nil {
				tokenStr = newToken
				c.Set("access_token", newToken)
				c.Header("X-New-Access-Token", newToken)
			}
		}

		// Re-validation avec clé secrète
		token, err = jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Méthode signature invalide")
			}
			return jwtSecret, nil
		})
		if err != nil || !token.Valid {
			c.Next()
			return
		}

		claims = token.Claims.(jwt.MapClaims)
		if userID, ok := claims["sub"].(string); ok {
			c.Set("user_id", userID)
		}

		c.Next()
	}
}
