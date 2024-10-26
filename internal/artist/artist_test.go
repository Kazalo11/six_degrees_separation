package artist

import (
	"reflect"
	"testing"

	"github.com/Kazalo11/six-degrees-seperation/internal/album"
	"github.com/zmb3/spotify/v2"
)

var (
	id1 = spotify.ID("id1")
	id2 = spotify.ID("id2")
	id3 = spotify.ID("id3")

	artist1 = album.Artist{
		ID: id1,
	}
	artist3 = album.Artist{
		ID: id3,
	}

	artist2 = album.Artist{
		ID:    id2,
		Name:  "name2",
		Songs: make([]string, 2),
	}
	feat1 = album.FeaturedArtistInfo{
		id2: artist2,
	}

	feat3 = album.FeaturedArtistInfo{
		id2: artist2,
	}
)

func TestMatchArtists(t *testing.T) {

	resp, _ := MatchArtists(feat1, feat3, id1, id3, nil)

	if len(resp) != 3 || resp == nil {
		t.Errorf("Did not return a response of length 3")
		return

	}

	expectedResponse := make([]album.Artist, 2)

	expectedResponse = append(expectedResponse, artist1)

	expectedResponse = append(expectedResponse, artist2)

	expectedResponse = append(expectedResponse, artist3)

	if reflect.DeepEqual(resp[0], artist2) {
		t.Errorf("Returned %v instead of artist2", resp)

	}

}
