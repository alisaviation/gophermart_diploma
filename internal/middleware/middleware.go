package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/skakunma/go-musthave-diploma-tpl/internal/config"
	jwtauth "github.com/skakunma/go-musthave-diploma-tpl/internal/jwt"
)

var authPaths = map[string]struct{}{
	"/api/user/orders":           {},
	"/api/user/profile":          {},
	"/api/user/balance/withdraw": {},
}

func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, exist := authPaths[c.Request.URL.Path]; exist {
			jwtToken, err := c.Cookie("jwt")
			if err != nil || jwtToken == "" {
				c.JSON(http.StatusUnauthorized, "You are not authorized")
				return
			}
			claims := &jwtauth.Claims{}
			token, err := jwt.ParseWithClaims(jwtToken, claims, func(t *jwt.Token) (interface{}, error) {
				return []byte(jwtauth.SecretKEY), nil
			})
			if err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Недействительный токен"})
				return
			}
			if !token.Valid {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Невалидный токен"})
				return
			}
			c.Set("user", claims)
			c.Next()
		}
		c.Next()
	}
}
