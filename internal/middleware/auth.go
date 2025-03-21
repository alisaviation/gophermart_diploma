package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rtmelsov/GopherMart/internal/config"
	"github.com/rtmelsov/GopherMart/internal/db"
	"github.com/rtmelsov/GopherMart/internal/models"
	"net/http"
	"strings"
	"time"
)

func Auth(conf config.ConfigI, db db.DBI) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := c.Cookie("Authorization")
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		tokenString = strings.TrimPrefix(tokenString, "Bearer ")
		// Parse the token
		token, err := jwt.ParseWithClaims(tokenString, &models.CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(conf.GetEnvVariables().Secret), nil
		})
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		claims := token.Claims.(*models.CustomClaims)

		if float64(time.Now().Unix()) > claims.Exp {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		user, localErr := db.GetUser(claims.Sub)
		if localErr != nil {
			c.AbortWithStatus(localErr.Code)
			return
		}
		if user == nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c.Set("userId", user.ID)
		c.Next()
	}
}
