package service

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"

	"github.com/TheHippo/podcastindex"
	"github.com/akhilrex/podgrab/model"
)

type SearchService interface {
	Query(q string) []*model.CommonSearchResultModel
}

type ItunesService struct {
}

const ITUNES_BASE = "https://itunes.apple.com"

func (service ItunesService) Query(q string) []*model.CommonSearchResultModel {
	url := fmt.Sprintf("%s/search?term=%s&entity=podcast", ITUNES_BASE, url.QueryEscape(q))

	body, _ := makeQuery(url)
	var response model.ItunesResponse
	json.Unmarshal(body, &response)

	var toReturn []*model.CommonSearchResultModel

	for _, obj := range response.Results {
		toReturn = append(toReturn, GetSearchFromItunes(obj))
	}

	return toReturn
}

type PodcastIndexService struct {
}

func (service PodcastIndexService) Query(q string) []*model.CommonSearchResultModel {
	key := os.Getenv("PODCASTINDEX_KEY")
	secret := os.Getenv("PODCASTINDEX_SECRET")
	if key == "" || secret == "" {
		fmt.Println("WARNING: PODCASTINDEX_KEY and/or PODCASTINDEX_SECRET not set; PodcastIndex search will be disabled")
		return nil
	}

	c := podcastindex.NewClient(key, secret)
	var toReturn []*model.CommonSearchResultModel
	podcasts, err := c.Search(q)
	if err != nil {
		fmt.Println("PodcastIndex search error:", err)
		return toReturn
	}

	for _, obj := range podcasts {
		toReturn = append(toReturn, GetSearchFromPodcastIndex(obj))
	}

	return toReturn
}
