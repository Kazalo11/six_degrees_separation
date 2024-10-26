package artist

import (
	"log"

	"github.com/Kazalo11/six-degrees-seperation/internal/album"
	"github.com/dominikbraun/graph"
	"github.com/zmb3/spotify/v2"
)

func UpsertGraph(feat1 album.FeaturedArtistInfo, feat2 album.FeaturedArtistInfo, startID spotify.ID, endID spotify.ID, curr graph.Graph[spotify.ID, album.Artist]) (graph.Graph[spotify.ID, album.Artist], error) {
	artistHash := func(a album.Artist) spotify.ID {
		return a.ID
	}
	var g graph.Graph[spotify.ID, album.Artist]
	if curr == nil {
		g = graph.New(artistHash, graph.Directed(), graph.Acyclic())
	} else {
		g = curr
	}

	startArtist := album.Artist{
		ID: startID,
	}

	endArtist := album.Artist{
		ID: endID,
	}

	err := g.AddVertex(startArtist)
	if err != nil {
		log.Printf("Vertex with ID: %s already exists", startID)
	}
	err2 := g.AddVertex(endArtist)

	if err2 != nil {
		log.Printf("Vertex with ID: %s already exists", endID)
	}

	for id, artist := range feat1 {
		g.AddVertex(artist)
		g.AddEdge(startID, id)

	}

	for id, artist := range feat2 {
		g.AddVertex(artist)
		g.AddEdge(startID, id)
	}

	path, err := graph.ShortestPath(g, startID, endID)

	if err != nil {
		log.Println("Can't find a path between them")
	}

	log.Println(path)

	return g, nil

}
