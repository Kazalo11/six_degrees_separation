package routes

import (
	"log"

	albumFuncs "github.com/Kazalo11/six-degrees-seperation/internal/album"
	"github.com/Kazalo11/six-degrees-seperation/internal/artist"
)

func iteration(featArr []albumFuncs.FeaturedArtistInfo, direction artist.Direction) {
	for _, feat := range featArr {
		upsertGraph(feat, direction)
	}

}

func upsertGraph(feat albumFuncs.FeaturedArtistInfo, direction artist.Direction) {
	for idx, newArtist := range feat {
		newFeat, err2 := featuredArtistInfo(idx.String())
		if direction == "forwards" {
			featForward = append(featForward, feat)
		} else {
			featBackward = append(featBackward, feat)
		}
		if err2 != nil {
			log.Printf("Could not find featured artists for %s due to err: %v", newArtist.Name, err2)
			continue
		}
		g = artist.UpsertGraph(newFeat, idx, direction, g)

	}
}
