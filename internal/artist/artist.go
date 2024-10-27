package artist

import (
	"errors"
	"log"

	"github.com/Kazalo11/six-degrees-seperation/internal/album"
	"github.com/dominikbraun/graph"
	"github.com/zmb3/spotify/v2"
)

type Direction string

const (
	forwards  Direction = "forwards"
	backwards Direction = "backwards"
)

func UpsertGraph(feat album.FeaturedArtistInfo, id spotify.ID, direction Direction, curr graph.Graph[spotify.ID, album.Artist]) graph.Graph[spotify.ID, album.Artist] {
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
		songData := make(map[string][]string)
		if direction == "forwards" {
			songData["from_songs"] = artist.Songs

		} else {
			songData["to_songs"] = artist.Songs
		}
		err = g.AddEdge(id, artistId, graph.EdgeData(songData))
		if err != nil {
			log.Printf("Can't add edge due to err: %v", err)
		}

	}
	return g

}

func MatchArtists(feat1 album.FeaturedArtistInfo, feat2 album.FeaturedArtistInfo, startID spotify.ID, endID spotify.ID, curr graph.Graph[spotify.ID, album.Artist]) ([]album.Artist, error) {
	g := UpsertGraph(feat1, startID, "forwards", curr)

	g2 := UpsertGraph(feat2, endID, "backwards", g)

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

		if songs, ok := data.(map[string][]string); ok {
			if from, ok := songs["from_songs"]; ok {
				artist.SongsFrom = from
			}
			if to, ok := songs["to_songs"]; ok {
				artist.SongsTo = to
			}

		}
		artists[i] = artist
	}
	final := len(ids) - 1
	artists[final], _ = g.Vertex(ids[final])

	return artists
}
