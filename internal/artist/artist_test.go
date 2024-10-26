package artist

import (
	"testing"

	"github.com/Kazalo11/six-degrees-seperation/internal/album"
	"github.com/stretchr/testify/assert"
	"github.com/zmb3/spotify/v2"
)

var (
	id1 = spotify.ID("id1")
	id2 = spotify.ID("id2")
	id3 = spotify.ID("id3")

	songs1 = make([]string, 2)
	songs2 = make([]string, 3)

	artist1 = album.Artist{
		ID: id1,
	}
	artist3 = album.Artist{
		ID: id3,
	}

	artist2 = album.Artist{
		ID:    id2,
		Name:  "name2",
		Songs: songs1,
	}

	artist2b = album.Artist{
		ID:    id2,
		Name:  "name2",
		Songs: songs2,
	}
	feat1 = album.FeaturedArtistInfo{
		id2: artist2,
	}

	feat3 = album.FeaturedArtistInfo{
		id2: artist2b,
	}
)

func init() {
	songs1[0] = "song1"
	songs1[1] = "song2"

	songs2[0] = "songs3"
	songs2[1] = "songs4"
	songs2[2] = "songs5"

}

func TestMatchArtists(t *testing.T) {

	resp, _ := MatchArtists(feat1, feat3, id1, id3, nil)

	if len(resp) != 3 || resp == nil {
		t.Errorf("Did not return a response of length 3")
		return

	}

	expectedArtist := album.Artist{
		ID:    id2,
		Name:  "name2",
		Songs: []string{"song1", "song2", "song3", "song4", "song5"},
	}

	expectedResponse := make([]album.Artist, 0)

	expectedResponse = append(expectedResponse, artist1)

	expectedResponse = append(expectedResponse, expectedArtist)

	expectedResponse = append(expectedResponse, artist3)

	if assert.Equal(t, resp, expectedResponse) {
		t.Errorf("Returned %v instead of artist2", resp)

	}

}
