package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/justlaputa/movie/pt"
	"github.com/mmcdole/gofeed"
)

const (
	//HDCRssURL rss link to retrieve hdc movies, you need to add passkey at the end
	HDCRssURL = "https://hdchina.org/torrentrss.php?rows=50&cat17=1&cat9=1&isize=1"

	hdcPasskeyEnvVar = "HDC_PASSKEY"
	//PutaoRssURL rss link to retrieve putao movies, no need to add passkey
	PutaoRssURL = "https://pt.sjtu.edu.cn/torrentrss.php?rows=50&cat401=1&cat402=1&cat403=1&sta1=1&sta3=1&isize=1"
)

//Handler the serverless entrypoint for `now`
func Handler(w http.ResponseWriter, r *http.Request) {
	hdcRssURLWithKey, err := addPasskey(HDCRssURL, hdcPasskeyEnvVar)
	if err != nil {
		log.Printf("failed to construct hdc rss url, %v", err)
		sendError(w, http.StatusInternalServerError)
		return
	}

	fp := gofeed.NewParser()
	hdcFeed, err := fp.ParseURL(hdcRssURLWithKey)
	if err != nil {
		sendError(w, http.StatusInternalServerError)
		return
	}

	log.Printf("got hdc feed from %s, contains %d items", hdcFeed.Title, len(hdcFeed.Items))

	updateItemsInDB(hdcFeed.Items, parseHDCItem)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{ \"status\": \"success\"}"))
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

func updateItemsInDB(items []*gofeed.Item, parseFunc PTItemParser) {
	for _, item := range items {
		movieInfo := parseFunc(item)
		log.Printf("parsed movie: %+v", movieInfo)
	}
}

//PTItemParser function definition for parsing a feed item from any pt rss
type PTItemParser func(*gofeed.Item) pt.MovieInfo

//ParseHDCItem parse an hdc rss feed item and return a MovieInfo
func parseHDCItem(item *gofeed.Item) pt.MovieInfo {
	info := pt.ParseHDCTitle(item.Title)

	if item.Link != "" {
		info.ID = extractID(item.Link)
	}

	if len(item.Enclosures) > 0 && item.Enclosures[0].Length != "" {
		if value, err := strconv.ParseUint(item.Enclosures[0].Length, 10, 64); err == nil {
			info.Size = pt.DigitalFileSize(value)
		}
	}

	return info
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
