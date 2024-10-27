package artist

import (
	"errors"
	"log"

	"github.com/Kazalo11/six-degrees-seperation/internal/album"
	"github.com/dominikbraun/graph"
	"github.com/zmb3/spotify/v2"
)

type Direction string

func UpsertGraph(feat album.FeaturedArtistInfo, id spotify.ID, curr graph.Graph[spotify.ID, album.Artist]) graph.Graph[spotify.ID, album.Artist] {
	artistHash := func(a album.Artist) spotify.ID {
		return a.ID
	}
	var g graph.Graph[spotify.ID, album.Artist]
	if curr == nil {
		g = graph.New(artistHash, graph.Acyclic())
	} else {
		g = curr
	}
	curr_artist := album.Artist{
		ID: id,
	}
	g.AddVertex(curr_artist)

	for artistId, artist := range feat {
		_, err := g.Vertex(artistId)
		if err != nil {
			g.AddVertex(artist)
		} else {
			log.Printf("Already found artist: %s for id: %s", artist.Name, id)
		}
		err = g.AddEdge(id, artistId, graph.EdgeData(artist.Songs))
		if err != nil {
			log.Printf("Can't add edge due to err: %v", err)
		}

	}
	return g

}

func MatchArtists(feat1 album.FeaturedArtistInfo, feat2 album.FeaturedArtistInfo, startID spotify.ID, endID spotify.ID, curr graph.Graph[spotify.ID, album.Artist]) ([]album.Artist, error) {
	g := UpsertGraph(feat1, startID, curr)

	g2 := UpsertGraph(feat2, endID, g)

	return GetShortestPath(g2, startID, endID)

}

func GetShortestPath(g graph.Graph[spotify.ID, album.Artist], startID spotify.ID, endID spotify.ID) ([]album.Artist, error) {
	pathIds, err := graph.ShortestPath(g, startID, endID)

	if err != nil || pathIds == nil {
		log.Println("Can't find a path between them")
		return nil, errors.New("can't find a path between them")
	}

	pathInfo := getPathArtistInfo(pathIds, g)

	return pathInfo, nil

}

func getPathArtistInfo(ids []spotify.ID, g graph.Graph[spotify.ID, album.Artist]) []album.Artist {
	artists := make([]album.Artist, len(ids))

	for i := 0; i < len(ids)-1; i++ {
		curr := ids[i]
		next := ids[i+1]

		edge, _ := g.Edge(curr, next)
		artist, _ := g.Vertex(curr)

		data := edge.Properties.Data

		if songs, ok := data.([]string); ok {
			artist.SongsConnection = songs

		}
		artists[i] = artist
	}
	final := len(ids) - 1
	artists[final], _ = g.Vertex(ids[final])

	return artists
}
