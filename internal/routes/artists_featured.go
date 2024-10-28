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

func getFeaturedArtists(c *gin.Context) {
	id := c.Param("id")

	if cachedData, found := cache.Get(id); found {
		log.Printf("Cache hit for artist: %s", id)
		c.JSON(http.StatusOK, cachedData)
		return
	}

	config := &clientcredentials.Config{
		ClientID:     os.Getenv("SPOTIFY_ID"),
		ClientSecret: os.Getenv("SPOTIFY_SECRET"),
		TokenURL:     spotifyauth.TokenURL,
	}
	token, err := config.Token(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to obtain authorization token"})
		return
	}

	client := spotify.New(auth.Client(c, token))

	params := url.Values{}
	params.Set("market", "US")
	params.Set("limit", "50")

	albums, err := client.GetArtistAlbums(c, spotify.ID(id), []spotify.AlbumType{1}, spotify.Limit(50), spotify.Market("US"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get albums due to err: %v", err)})
		return
	}

	featuredArtists := make(albumFuncs.FeaturedArtistInfo)
	var wg sync.WaitGroup

	resultChan := make(chan albumFuncs.FeaturedArtistInfo)

	batchSize := 20

	for i := 0; i < len(albums.Albums); i += batchSize {
		end := i + batchSize
		if end > len(albums.Albums) {
			end = len(albums.Albums)
		}

		albumBatch := albums.Albums[i:end]

		albumIds := albumFuncs.GetAlbumIDs(albumBatch)

		fullAlbums, err := client.GetAlbums(c, albumIds)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get songs from album due to err: %v", err)})
			return
		}

		for _, fullAlbum := range fullAlbums {
			wg.Add(1)
			go func(albumId spotify.ID) {
				defer wg.Done()
				featuredArtists := albumFuncs.GetArtistsFromAlbum(fullAlbum, id)
				resultChan <- featuredArtists
			}(fullAlbum.ID)

		}

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

	cache.Set(id, featuredArtists, 1)
	cache.Wait()
	log.Println("Wrote to the cache")

	c.JSON(http.StatusOK, featuredArtists)

}
