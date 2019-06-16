package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/mmcdole/gofeed"
	"github.com/moviegeek/pt"
)

const (
	//HDCRssURL rss link to retrieve hdc movies, you need to add passkey at the end
	HDCRssURL = "https://hdchina.org/torrentrss.php?rows=50&cat17=1&cat9=1&isize=1"
	//HDCSiteName name of the source from hdc
	HDCSiteName = "HDChina"
	//PutaoSiteName name of the soruce from putao
	PutaoSiteName = "Putao"

	hdcPasskeyEnvVar = "HDC_PASSKEY"
	//PutaoRssURL rss link to retrieve putao movies, no need to add passkey
	PutaoRssURL = "https://pt.sjtu.edu.cn/torrentrss.php?rows=50&cat401=1&cat402=1&cat403=1&sta1=1&sta3=1&isize=1"
)

//PTMovie add some extra fields to the pt.MovieInfo
type PTMovie struct {
	pt.MovieInfo
	ID       string
	SiteName string
}

//Handler the serverless entrypoint for `now`
func Handler(w http.ResponseWriter, r *http.Request) {
	hdcRssURLWithKey, err := addPasskey(HDCRssURL, hdcPasskeyEnvVar)
	if err != nil {
		log.Printf("failed to construct hdc rss url, %v", err)
		sendError(w, http.StatusInternalServerError)
		return
	}

	fp := gofeed.NewParser()
	hdcMovies := []PTMovie{}
	hdcFeed, err := fp.ParseURL(hdcRssURLWithKey)
	if err != nil {
		log.Printf("failed to get HDC Rss, %v", err)
	} else {
		log.Printf("got hdc feed from %s, contains %d items", hdcFeed.Title, len(hdcFeed.Items))
		hdcMovies = parseFeedItems(hdcFeed.Items, HDCSiteName)
	}

	putaoFeed, err := fp.ParseURL(PutaoRssURL)
	putaoMovies := []PTMovie{}
	if err != nil {
		log.Printf("failed to get Putao Rss, %v", err)
	} else {
		log.Printf("got putao feed from %s, contains %d items", putaoFeed.Title, len(putaoFeed.Items))
		putaoMovies = parseFeedItems(putaoFeed.Items, PutaoSiteName)
	}

	movies := append(hdcMovies, putaoMovies...)

	data, err := json.Marshal(movies)
	if err != nil {
		sendError(w, http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func addPasskey(url, envVar string) (string, error) {
	if value := os.Getenv(envVar); value != "" {
		return fmt.Sprintf("%s&passkey=%s", url, value), nil
	}
	return "", fmt.Errorf("can not find environment variable %s", envVar)
}

func sendError(w http.ResponseWriter, statusCode int) {
	w.WriteHeader(statusCode)
}

func parseFeedItems(items []*gofeed.Item, siteName string) []PTMovie {
	movies := []PTMovie{}

	for _, item := range items {
		info := pt.ParseTitle(item.Title)
		movie := PTMovie{info, "", siteName}

		if item.Link != "" {
			movie.ID = extractID(item.Link)
		}

		if len(item.Enclosures) > 0 && item.Enclosures[0].Length != "" {
			if value, err := strconv.ParseUint(item.Enclosures[0].Length, 10, 64); err == nil {
				info.Size = pt.DigitalFileSize(value)
			}
		}

		movies = append(movies, movie)
	}

	return movies
}

func extractID(url string) string {
	id := ""
	i := strings.LastIndex(url, "id=")
	if i > -1 {
		id = url[i+3:]
		id = strings.TrimSpace(id)
	}
	return id
}
