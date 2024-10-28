package routes

import (
	"fmt"
	"net/http"

	"github.com/Kazalo11/six-degrees-seperation/internal/artist"
	"github.com/gin-gonic/gin"
	spotify "github.com/zmb3/spotify/v2"
)

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
