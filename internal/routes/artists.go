package routes

import (
	"fmt"
	"log"
	"net/http"

	albumFuncs "github.com/Kazalo11/six-degrees-seperation/internal/album"
	"github.com/Kazalo11/six-degrees-seperation/internal/artist"
	"github.com/dgraph-io/ristretto"
	"github.com/dominikbraun/graph"
	"github.com/gin-gonic/gin"
	spotify "github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
)

var (
	redirectURL  = "http://localhost:8080/v1/artist/callback"
	auth         = spotifyauth.New(spotifyauth.WithRedirectURL(redirectURL))
	g            graph.Graph[spotify.ID, albumFuncs.Artist]
	cache_config = &ristretto.Config{
		NumCounters: 1e7,
		MaxCost:     1 << 30,
		BufferItems: 64,
		OnReject: func(item *ristretto.Item) {
			fmt.Printf("Rejected item = %+v\n", item)
		},
	}
	cache    *ristretto.Cache
	featCurr = make([]albumFuncs.FeaturedArtistInfo, 0)
	featPrev = make([]albumFuncs.FeaturedArtistInfo, 0)
)

func init() {
	var initError error
	cache, initError = ristretto.NewCache(cache_config)
	log.Println("Cache initalised")
	if initError != nil {
		log.Printf("Could not initalise cache  due to err: %v ", initError)
	}
}

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

	g = artist.UpsertGraph(feat1, spotify.ID(id1), nil)
	g = artist.UpsertGraph(feat2, spotify.ID(id2), g)

	path, _ := artist.GetShortestPath(g, spotify.ID(id1), spotify.ID(id2))

	if path != nil {
		c.JSON(http.StatusOK, path)
		return
	}

	upsertGraph(feat1)

	upsertGraph(feat2)

	path, _ = artist.GetShortestPath(g, spotify.ID(id1), spotify.ID(id2))

	if path != nil {
		c.JSON(http.StatusOK, path)
		return
	}

	maxIterations := 10

	for i := 0; i < maxIterations; i++ {
		iteration(featPrev)

		path, _ = artist.GetShortestPath(g, spotify.ID(id1), spotify.ID(id2))

		if path != nil {
			c.JSON(http.StatusOK, path)
			return
		}

		featPrev = featCurr
		featCurr = nil

	}

}
