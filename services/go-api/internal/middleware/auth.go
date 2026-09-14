package middleware

import (
	"net/http"

	appauth "github.com/Jennifer-ra/accounting/services/go-api/internal/auth"
	"github.com/Jennifer-ra/accounting/services/go-api/internal/response"
	"github.com/Jennifer-ra/accounting/services/go-api/internal/service"
	"github.com/gin-gonic/gin"
)

const userIDKey = "userID"

func Auth(authSvc *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := appauth.Bearer(c.GetHeader("Authorization"))
		if err != nil {
			response.Error(c, http.StatusUnauthorized, "unauthorized", "missing token")
			c.Abort()
			return
		}
		user, err := authSvc.UserFromToken(token)
		if err != nil {
			response.Error(c, http.StatusUnauthorized, "unauthorized", "invalid token")
			c.Abort()
			return
		}
		c.Set(userIDKey, user.ID)
		c.Next()
	}
}

func UserID(c *gin.Context) uint {
	value, exists := c.Get(userIDKey)
	if !exists {
		return 0
	}
	id, _ := value.(uint)
	return id
}
