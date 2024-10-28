package routes

import (
	"fmt"
	"log"

	albumFuncs "github.com/Kazalo11/six-degrees-seperation/internal/album"
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
