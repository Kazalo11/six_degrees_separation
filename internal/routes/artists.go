package routes

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"sync"

	albumFuncs "github.com/Kazalo11/six-degrees-seperation/internal/album"
	"github.com/gin-gonic/gin"
	spotify "github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2/clientcredentials"
)

var (
	redirectURL = "http://localhost:8080/v1/artist/callback"
	auth        = spotifyauth.New(spotifyauth.WithRedirectURL(redirectURL))
)

func ArtistRoutes(superRoute *gin.RouterGroup) {
	artistRouter := superRoute.Group("/artist")
	{
		artistRouter.GET("/:id/features", getFeaturedArtists)
	}

}

func getFeaturedArtists(c *gin.Context) {
	id := c.Param("id")

	config := &clientcredentials.Config{
		ClientID:     os.Getenv("SPOTIFY_ID"),
		ClientSecret: os.Getenv("SPOTIFY_SECRET"),
		TokenURL:     spotifyauth.TokenURL,
	}
	token, err := config.Token(c)
	if err != nil {
		log.Printf("Failed to get token: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to obtain authorization token"})
		return
	}

	client := spotify.New(auth.Client(c, token))

	params := url.Values{}
	params.Set("market", "US")
	params.Set("limit", "50")

	albums, err := client.GetArtistAlbums(c, spotify.ID(id), []spotify.AlbumType{1}, spotify.Limit(50), spotify.Market("US"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get albums"})
		return
	}

	featuredArtists := make(albumFuncs.FeaturedArtistInfo)
	var wg sync.WaitGroup

	resultChan := make(chan albumFuncs.FeaturedArtistInfo)

	for _, album := range albums.Albums {
		wg.Add(1)
		go func(albumId spotify.ID) {
			defer wg.Done()
			fullAlbum, err := client.GetAlbum(c, album.ID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get songs from album %s: %v", album.ID, err)})
				return
			}
			featuredArtist := albumFuncs.GetArtistsFromAlbum(fullAlbum, id)
			resultChan <- featuredArtist

		}(album.ID)
	}
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	go func() {
		for featuredArtist := range resultChan {
			featuredArtists = albumFuncs.MergeEntries(featuredArtist, featuredArtists)
		}
	}()

	wg.Wait()
	c.JSON(http.StatusOK, featuredArtists)

}
