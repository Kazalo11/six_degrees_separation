package routes

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	albumFuncs "github.com/Kazalo11/six-degrees-seperation/internal/album"
)

func featuredArtistInfo(id string) (albumFuncs.FeaturedArtistInfo, error) {
	var feat albumFuncs.FeaturedArtistInfo
	domain := os.Getenv("DOMAIN")
	resp1, err := http.Get(fmt.Sprintf("%s/v1/artist/%s/features", domain, id))
	if err != nil {
		log.Printf("Failed to get features for artist: %s", id)
		return nil, err
	}
	body, err := io.ReadAll(resp1.Body)
	if err != nil {
		log.Printf("Failed to read response body for artist: %s", id)
		return nil, err
	}

	err = json.Unmarshal(body, &feat)
	if err != nil {
		log.Printf("Failed to decode json for artist: %s", id)
		return feat, err
	}
	return feat, nil

}
