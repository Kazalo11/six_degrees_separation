package routes

import (
	"log"

	albumFuncs "github.com/Kazalo11/six-degrees-seperation/internal/album"
	"github.com/Kazalo11/six-degrees-seperation/internal/artist"
)

func iteration(featArr []albumFuncs.FeaturedArtistInfo) {
	for _, feat := range featArr {
		upsertGraph(feat)
	}

}

func upsertGraph(feat albumFuncs.FeaturedArtistInfo) {
	for idx, newArtist := range feat {
		newFeat, err2 := featuredArtistInfo(idx.String())
		featCurr = append(featCurr, newFeat)
		if err2 != nil {
			log.Printf("Could not find featured artists for %s due to err: %v", newArtist.Name, err2)
			continue
		}
		g = artist.UpsertGraph(newFeat, idx, g)

	}
}
