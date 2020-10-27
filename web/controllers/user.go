package controllers

import (
	"fmt"
	"net/http"

	oauth1Login "github.com/dghubble/gologin/v2/oauth1"
	"github.com/dghubble/gologin/v2/twitter"
	"github.com/dghubble/oauth1"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type UserController struct {
	oauth1Config *oauth1.Config
}

func NewUserController(oauth1Config *oauth1.Config) *UserController {
	return &UserController{oauth1Config}
}

func (controller UserController) Login(c *gin.Context) {
	gin.WrapH(
		twitter.LoginHandler(controller.oauth1Config, nil),
	)(c)
}

func (controller UserController) LoginCallback(c *gin.Context) {
	issueSession := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		twitterUser, err := twitter.UserFromContext(ctx)
		accessToken, accessSecret, err := oauth1Login.AccessTokenFromContext(ctx)
		fmt.Println(accessToken, accessSecret)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		session := sessions.Default(c)
		session.Set("user_id", twitterUser.ID)
		session.Set("user_screen_name", twitterUser.ScreenName)
		session.Save()
		http.Redirect(w, req, "/", http.StatusFound)
	}

	gin.WrapH(
		twitter.CallbackHandler(controller.oauth1Config, http.HandlerFunc(issueSession), nil),
	)(c)
}

func (controller UserController) Index(c *gin.Context) {
	userID := c.MustGet("user_id").(int64)
	userScreenName := c.MustGet("user_screen_name").(string)
	// show list of accounts
	c.HTML(http.StatusOK, "user.tmpl", gin.H{
		"userID":         userID,
		"userScreenName": userScreenName,
	})
}

/*
func Signin(c *gin.Context) {
	session := sessions.Default(c)
	session.Set("user_id", 1)
	session.Set("user_email", "demo@demo.com")
	session.Set("user_username", "demo")
	session.Save()

	c.JSON(http.StatusOK, gin.H{"message": "User signed in", "user": "demo"})
}

func Signout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	session.Save()
	c.JSON(http.StatusOK, gin.H{"message": "Signed out..."})
}

*/
