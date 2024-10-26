package artist

import (
	"log"

	"github.com/Kazalo11/six-degrees-seperation/internal/album"
	"github.com/dominikbraun/graph"
	"github.com/zmb3/spotify/v2"
)

func UpsertGraph(feat1 album.FeaturedArtistInfo, feat2 album.FeaturedArtistInfo, startID spotify.ID, endID spotify.ID, curr graph.Graph[spotify.ID, album.Artist]) ([]album.Artist, error) {
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
		err := g.AddVertex(artist)
		if err != nil {
			art, err := g.Vertex(id)
			if err != nil {
				log.Println("Failed to add vertex with id: %s due to err %v", id, err)
			}
		}
		g.AddEdge(endID, id)
	}

	pathIds, err := graph.ShortestPath(g, startID, endID)

	if err != nil {
		log.Println("Can't find a path between them")
	}

	log.Println(pathIds)

	pathInfo := getPathArtistInfo(pathIds, g)

	return pathInfo, nil

}

func getPathArtistInfo(ids []spotify.ID, g graph.Graph[spotify.ID, album.Artist]) []album.Artist {
	artists := make([]album.Artist, len(ids)-1)

	for idx, id := range ids {
		if idx == 0 {
			continue
		}
		artist, err := g.Vertex(id)
		if err != nil {
			log.Printf("Can't find artist with id: %s", id)
		}
		artists[idx-1] = artist
	}

	return artists
}
