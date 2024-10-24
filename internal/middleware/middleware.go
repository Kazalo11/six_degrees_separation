package middleware

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2/clientcredentials"
)

func AuthorizeRequest() gin.HandlerFunc {
	return func(c *gin.Context) {
		config := &clientcredentials.Config{
			ClientID:     os.Getenv("SPOTIFY_ID"),
			ClientSecret: os.Getenv("SPOTIFY_SECRET"),
			TokenURL:     spotifyauth.TokenURL,
		}
		token, err := config.Token(c)
		if err != nil {
			log.Printf("Failed to get token: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to obtain authorization token"})
			c.Abort()
			return
		}
		c.Set("spotify_token", token)

		c.Next()
	}
}
