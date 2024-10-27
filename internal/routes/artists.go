package routes

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"sync"

	albumFuncs "github.com/Kazalo11/six-degrees-seperation/internal/album"
	"github.com/Kazalo11/six-degrees-seperation/internal/artist"
	"github.com/dominikbraun/graph"
	"github.com/gin-gonic/gin"
	spotify "github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2/clientcredentials"
)

var (
	redirectURL = "http://localhost:8080/v1/artist/callback"
	auth        = spotifyauth.New(spotifyauth.WithRedirectURL(redirectURL))
	g           graph.Graph[spotify.ID, albumFuncs.Artist]
)

func ArtistRoutes(superRoute *gin.RouterGroup) {
	artistRouter := superRoute.Group("/artist")
	{
		artistRouter.GET("/:id/features", getFeaturedArtists)
		artistRouter.GET("/connect/:id1/:id2", connectArtists)
	}

}

func connectArtists(c *gin.Context) {
	id1 := c.Param("id1")
	id2 := c.Param("id2")

	feat1, err := featuredArtistInfo(id1)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Failed to get artists for id %s due to err %v", id1, err)})
		return
	}

	feat2, err := featuredArtistInfo(id2)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Failed to get artists for id %s due to err %v", id2, err)})
		return
	}

	g = artist.UpsertGraph(feat1, spotify.ID(id1), "forwards", nil)
	g = artist.UpsertGraph(feat2, spotify.ID(id2), "backwards", g)

	path, _ := artist.GetShortestPath(g, spotify.ID(id1), spotify.ID(id2))

	if path != nil {
		c.JSON(http.StatusOK, path)
		return
	}

	for idx, newArtist := range feat1 {
		newFeat, err := featuredArtistInfo(idx.String())
		if err != nil {
			log.Printf("Could not find featured artists for %s due to err: %v", newArtist.Name, err)
			continue
		}
		g = artist.UpsertGraph(newFeat, idx, "forwards", g)

	}

	for idx, newArtist := range feat2 {
		newFeat, err2 := featuredArtistInfo(idx.String())
		if err2 != nil {
			log.Printf("Could not find featured artists for %s due to err: %v", newArtist.Name, err2)
			continue
		}
		g = artist.UpsertGraph(newFeat, idx, "backwards", g)

	}

	path, _ = artist.GetShortestPath(g, spotify.ID(id1), spotify.ID(id2))

	if path != nil {
		c.JSON(http.StatusOK, path)
		return
	}

}

func featuredArtistInfo(id string) (albumFuncs.FeaturedArtistInfo, error) {
	var feat albumFuncs.FeaturedArtistInfo
	domain := os.Getenv("DOMAIN")
	resp1, err := http.Get(fmt.Sprintf("%s/v1/artist/%s/features", domain, id))
	if err != nil {
		log.Printf("Failed to get features for artist: %s", id)
		return nil, err
	}
	body, err := io.ReadAll(resp1.Body)
	if err != nil {
		log.Printf("Failed to read response body for artist: %s", id)
		return feat, err
	}

	err = json.Unmarshal(body, &feat)
	if err != nil {
		log.Printf("Failed to decode json for artist: %s", id)
		return feat, err
	}
	return feat, nil

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
	c.JSON(http.StatusOK, featuredArtists)

}
