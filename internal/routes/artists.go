package routes

import (
	"encoding/json"
	"fmt"
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
	wg          sync.WaitGroup
	mu          sync.Mutex
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

	mu.Lock()
	g = artist.UpsertGraph(feat1, spotify.ID(id1), "forwards", nil)
	g = artist.UpsertGraph(feat2, spotify.ID(id2), "backwards", g)
	mu.Unlock()

	path, _ := artist.GetShortestPath(g, spotify.ID(id1), spotify.ID(id2))

	if path != nil {
		c.JSON(http.StatusOK, path)
		return
	}

	process := func(id spotify.ID, direction artist.Direction) {
		defer wg.Done()
		feat, err := featuredArtistInfo(id.String())
		if err != nil {
			log.Printf("Failed to get artists for id %s due to err %v \n", id, err)
			return
		}
		mu.Lock()
		g = artist.UpsertGraph(feat, id, direction, g)
		mu.Unlock()

	}

	for idx := range feat1 {
		wg.Add(1)
		go process(idx, "forwards")
	}

	for idx := range feat2 {
		wg.Add(1)
		go process(idx, "backwards")
	}

	wg.Wait()

	path, _ = artist.GetShortestPath(g, spotify.ID(id1), spotify.ID(id2))

	if path != nil {
		c.JSON(http.StatusOK, path)
		return
	}
	c.JSON(http.StatusBadRequest, "Could not find path between ids")

}

func featuredArtistInfo(id string) (albumFuncs.FeaturedArtistInfo, error) {
	var feat albumFuncs.FeaturedArtistInfo
	domain := os.Getenv("DOMAIN")
	resp1, err := http.Get(fmt.Sprintf("%s/v1/artist/%s/features", domain, id))
	if err != nil {
		log.Printf("Failed to get features for artist: %s", id)
		return nil, err
	}

	err = json.NewDecoder(resp1.Body).Decode(&feat)
	if err != nil {
		log.Printf("Failed to decode json for artist: %s", id)
		return nil, err
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
