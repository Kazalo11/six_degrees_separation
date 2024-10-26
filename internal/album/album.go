package album

import (
	"maps"

	spotify "github.com/zmb3/spotify/v2"
)

type Artist struct {
	Name      string
	Songs     []string
	ID        spotify.ID
	SongsFrom []string
	SongsTo   []string
}

type FeaturedArtistInfo map[spotify.ID]Artist

func GetArtistsFromAlbum(fullAlbum *spotify.FullAlbum, originalArtistID string) FeaturedArtistInfo {
	tracks := fullAlbum.Tracks.Tracks
	response := make(FeaturedArtistInfo)

	for _, track := range tracks {
		artists := track.Artists
		for _, artist := range artists {
			if artist.ID == spotify.ID(originalArtistID) {
				continue
			}
			existingArtist, ok := response[artist.ID]
			if ok {
				existingArtist.Songs = append(existingArtist.Songs, track.Name)
				response[artist.ID] = existingArtist

			} else {
				response[artist.ID] = Artist{
					Name:  artist.Name,
					ID:    artist.ID,
					Songs: []string{track.Name},
				}
			}
		}

	}

	return response

}

func MergeEntries(new FeaturedArtistInfo, curr FeaturedArtistInfo) FeaturedArtistInfo {
	response := maps.Clone(curr)
	for key, value := range new {
		info, exists := response[key]
		if exists {
			info.Songs = removeDuplicates(append(info.Songs, new[key].Songs...))
			response[key] = info
		} else {
			response[key] = value
		}
	}

	return response
}

func removeDuplicates(input []string) []string {
	seen := make(map[string]struct{})
	var result []string

	for _, item := range input {
		if _, ok := seen[item]; !ok {
			seen[item] = struct{}{}
			result = append(result, item)
		}
	}

	return result
}
