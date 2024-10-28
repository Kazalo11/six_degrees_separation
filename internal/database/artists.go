package database

import (
	"database/sql"
	"errors"
	"log"

	"github.com/Kazalo11/six-degrees-seperation/internal/album"
	"github.com/lib/pq"
)

func WriteFeaturedArtists(db *sql.DB, id string, feat album.FeaturedArtistInfo) (string, error) {
	var overall_error error
	for _, artist := range feat {
		_, err := db.Exec(
			"INSERT INTO featured_artists (id, name, songs, featured_artist_id) VALUES ($1, $2, $3, $4)",
			id,
			artist.Name,
			pq.Array(artist.Songs), // Use pq.Array to insert a slice as a PostgreSQL array
			artist.ID,
		)
		if err != nil {
			log.Printf("Error adding featured artist into table for id: %s and featured_artist id: %s, error: %v", id, artist.ID, err)
			overall_error = errors.Join(overall_error, err)
		}
	}
	return id, overall_error

}
