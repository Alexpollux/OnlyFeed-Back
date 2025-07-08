package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ArthurDelaporte/OnlyFeed-Back/internal/logs"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		route := c.FullPath()

		authHeader := c.GetHeader("Authorization")
		refreshToken := c.GetHeader("X-Refresh-Token")

		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token requis"})
			logs.LogJSON("WARN", "Token required", map[string]interface{}{
				"route": route,
			})
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		jwtSecret := []byte(os.Getenv("JWT_SECRET"))

		// Parse the token WITHOUT validating expiration (we handle it manually)
		token, _, err := new(jwt.Parser).ParseUnverified(tokenStr, jwt.MapClaims{})
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token illisible", "details": err.Error()})
			logs.LogJSON("ERROR", "Illegible token", map[string]interface{}{
				"error": err.Error(),
				"route": route,
				"extra": fmt.Sprintf("Illegible token : %#v", tokenStr),
			})
			return
		}

		claims := token.Claims.(jwt.MapClaims)
		expFloat, ok := claims["exp"].(float64)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token invalide ou champ exp manquant"})
			logs.LogJSON("ERROR", "Invalid token or missing “exp” field", map[string]interface{}{
				"route": route,
				"extra": fmt.Sprintf("Invalid token or missing “exp” field : %v", claims),
			})
			return
		}

		exp := int64(expFloat)
		now := time.Now().Unix()

		// Token expiré + refresh disponible → essayer de rafraîchir
		if now > exp && refreshToken != "" {
			newToken, err := refreshAccessToken(refreshToken)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token expiré et refresh échoué", "details": err.Error()})
				logs.LogJSON("ERROR", "Token expired and refresh failed", map[string]interface{}{
					"error": err.Error(),
					"route": route,
				})
				return
			}
			tokenStr = newToken
			c.Set("access_token", newToken)
			c.Header("X-New-Access-Token", newToken)
		}

		// Maintenant on parse + valide le token avec la clé
		token, err = jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Méthode signature invalide")
			}
			return jwtSecret, nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token final invalide", "details": err.Error()})
			logs.LogJSON("ERROR", "Invalid final token", map[string]interface{}{
				"error": err.Error(),
				"route": route,
			})
			return
		}

		claims = token.Claims.(jwt.MapClaims)
		userID, ok := claims["sub"].(string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "User ID manquant"})
			logs.LogJSON("ERROR", "User ID missing", map[string]interface{}{
				"route": route,
				"extra": fmt.Sprintf("User ID missing: %v", claims),
			})
			return
		}
		userEmail, ok := claims["email"].(string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "User Email manquant"})
			logs.LogJSON("ERROR", "User Email missing", map[string]interface{}{
				"route": route,
				"extra": fmt.Sprintf("User Email missing : %v", claims),
			})
			return
		}

		c.Set("user_id", userID)
		c.Set("user_email", userEmail)
		c.Next()
	}
}

func refreshAccessToken(refreshToken string) (string, error) {
	supabaseBaseURL := os.Getenv("NEXT_PUBLIC_SUPABASE_URL")
	anonKey := os.Getenv("SUPABASE_ANON_KEY")

	payload := map[string]string{"refresh_token": refreshToken}
	jsonBody, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", supabaseBaseURL+"/auth/v1/token?grant_type=refresh_token", bytes.NewBuffer(jsonBody))
	req.Header.Set("apikey", anonKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("appel Supabase échoué: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("erreur Supabase: %s", body)
	}

	var response struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("parse JSON échoué: %v", err)
	}

	return response.AccessToken, nil
}
