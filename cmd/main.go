package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	twitterOAuth1 "github.com/dghubble/oauth1/twitter"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"github.com/grevych/rub/web/controllers"
	"github.com/grevych/rub/web/middlewares"
	_ "github.com/joho/godotenv/autoload"
	"golang.org/x/crypto/acme/autocert"
)

var (
	RedisHost                = os.Getenv("REDIS_HOST")
	RedisPort                = os.Getenv("REDIS_PORT")
	TwitterConsumerKey       = os.Getenv("TWITTER_CONSUMER_KEY")
	TwitterConsumerSecret    = os.Getenv("TWITTER_CONSUMER_SECRET")
	TwitterAccessToken       = os.Getenv("TWITTER_ACCESS_TOKEN")
	TwitterAccessTokenSecret = os.Getenv("TWITTER_ACCESS_TOKEN_SECRET")
)

func main() {
	router := gin.Default()
	router.Static("/assets", "./web/assets")
	router.LoadHTMLGlob("./web/templates/*.tmpl")

	store, _ := redis.NewStore(
		10, "tcp", RedisHost+":"+RedisPort, "", []byte("secret"),
	)
	router.Use(sessions.Sessions("mysession", store))

	oauth1Config := &oauth1.Config{
		ConsumerKey:    TwitterConsumerKey,
		ConsumerSecret: TwitterConsumerSecret,
		CallbackURL:    "http://localhost:8080/twitter/callback",
		Endpoint:       twitterOAuth1.AuthorizeEndpoint,
	}
	userController := controllers.NewUserController(oauth1Config)

	router.GET("/twitter/login", userController.Login)
	router.GET("/twitter/callback", userController.LoginCallback)

	auth := router.Group("/")
	auth.Use(middlewares.AuthRequired())
	{
		auth.GET("/", userController.Index)
	}

	config := oauth1.NewConfig(TwitterConsumerKey, TwitterConsumerSecret)
	twitterToken := oauth1.NewToken(TwitterAccessToken, TwitterAccessTokenSecret)
	httpClient := config.Client(oauth1.NoContext, twitterToken)
	client := twitter.NewClient(httpClient)

	tweets, resp, _ := client.Timelines.HomeTimeline(&twitter.HomeTimelineParams{
		Count: 20,
	})

	fmt.Println(tweets)
	fmt.Println(resp)

	// Listen and serve on 0.0.0.0:8080
	// router.Run(":8080")

	// Initializing the server in a goroutine so that
	// it won't block the graceful shutdown handling below
	// go http.ListenAndServe(":http", http.HandlerFunc(redirect))
	// go serve(router)

	srv := &http.Server{
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	// go serveTLS(srv)
	go serve(srv)

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be catch, so don't need add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}

func serveTLS(srv *http.Server) {
	srv.Addr = ":81"
	listener := autocert.NewListener("reporta1bot.mx", "reporta1bot.com")
	if err := srv.Serve(listener); err != nil && err != http.ErrServerClosed {
		log.Fatalf("listen: %s\n", err)
	}
}

func serve(srv *http.Server) {
	srv.Addr = ":8080"
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("listen: %s\n", err)
	}
}

// func redirect(w http.ResponseWriter, req *http.Request) {
// 	target := "https://" + req.Host + req.URL.Path
//
// 	if len(req.URL.RawQuery) > 0 {
// 		target += "?" + req.URL.RawQuery
// 	}
//
// 	http.Redirect(w, req, target, http.StatusTemporaryRedirect)
// }
