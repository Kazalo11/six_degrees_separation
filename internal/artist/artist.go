package artist

import (
	"log"

	"github.com/Kazalo11/six-degrees-seperation/internal/album"
	"github.com/dominikbraun/graph"
	"github.com/zmb3/spotify/v2"
)

func MatchArtists(feat1 album.FeaturedArtistInfo, feat2 album.FeaturedArtistInfo, startID spotify.ID, endID spotify.ID, curr graph.Graph[spotify.ID, album.Artist]) ([]album.Artist, graph.Graph[spotify.ID, album.Artist]) {
	artistHash := func(a album.Artist) spotify.ID {
		return a.ID
	}
	var g graph.Graph[spotify.ID, album.Artist]
	if curr == nil {
		g = graph.New(artistHash, graph.Acyclic())
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
		_, err := g.Vertex(id)
		if err != nil {
			log.Printf("Adding artist: %s \n", artist.Name)
			g.AddVertex(artist)
		} else {
			log.Printf("Already found artist: %s for id: %s", artist.Name, startID)

		}
		songData := make(map[string][]string)
		songData["from_songs"] = artist.Songs
		err = g.AddEdge(startID, id, graph.EdgeData(songData))
		if err != nil {
			log.Printf("Can't add edge due to err: %v", err)
		}

	}

	for id, artist := range feat2 {
		_, err := g.Vertex(id)
		if err != nil {
			log.Printf("Adding artist: %s \n", artist.Name)
			g.AddVertex(artist)
		} else {
			log.Printf("Already found artist: %s for id: %s", artist.Name, endID)

		}
		songData := make(map[string][]string)
		songData["to_songs"] = artist.Songs
		err = g.AddEdge(endID, id, graph.EdgeData(songData))
		if err != nil {
			log.Printf("Can't add edge due to err: %v", err)
		}

	}

	pathIds, err := graph.ShortestPath(g, startID, endID)

	if err != nil || pathIds == nil {
		log.Println("Can't find a path between them")
		return nil, g
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
