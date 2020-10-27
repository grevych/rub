package middlewares

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
			c.Set("user_screen_name", session.Get("user_screen_name"))
			c.Next()
			return
		}

		c.HTML(http.StatusOK, "index.tmpl", gin.H{})
		c.Abort()
	}
}
