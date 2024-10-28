package database

import (
	"database/sql"
	"errors"
	"log"

	"github.com/Kazalo11/six-degrees-seperation/internal/album"
	"github.com/lib/pq"
	"github.com/zmb3/spotify/v2"
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

func GetFeaturedArtists(db *sql.DB, id string) (album.FeaturedArtistInfo, error) {
	rows, err := db.Query("SELECT * FROM featured_artists WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	feat := make(album.FeaturedArtistInfo)
	for rows.Next() {
		var art album.Artist
		var string_id spotify.ID
		if err := rows.Scan(&string_id, &art.Name, pq.Array(&art.Songs), &art.ID); err != nil {
			return feat, err

		}
		feat[art.ID] = art
	}
	if err = rows.Err(); err != nil {
		return feat, err
	}
	return feat, nil
}
