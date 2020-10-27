package web

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		if sessionID := session.Get("user_id"); sessionID != nil {
			c.Set("user_id", sessionID)
			return c.Next()
		}

		c.HTML(http.StatusOK, "index.tmpl", gin.H{})
		c.Abort()
	}
}
