package rss

import (
	"encoding/xml"
	"net/http"

	"github.com/paulrosania/go-charset/charset"
)

//Channel struct for RSS
type Channel struct {
	Title         string `xml:"title"`
	Link          string `xml:"link"`
	Description   string `xml:"description"`
	Language      string `xml:"language"`
	LastBuildDate Date   `xml:"lastBuildDate"`
	Item          []Item `xml:"item"`
}

//ItemEnclosure struct for each Item Enclosure
type ItemEnclosure struct {
	URL  string `xml:"url,attr"`
	Type string `xml:"type,attr"`
}

//Item struct for each Item in the Channel
type Item struct {
	Title       string          `xml:"title"`
	Link        string          `xml:"link"`
	Comments    string          `xml:"comments"`
	PubDate     Date            `xml:"pubDate"`
	GUID        string          `xml:"guid"`
	Category    []string        `xml:"category"`
	Enclosure   []ItemEnclosure `xml:"enclosure"`
	Description string          `xml:"description"`
	Author      string          `xml:"author"`
	Content     string          `xml:"content"`
	FullText    string          `xml:"full-text"`
}

//Regular parses regular feeds
func Regular(resp *http.Response) (*Channel, error) {
	defer resp.Body.Close()
	xmlDecoder := xml.NewDecoder(resp.Body)
	xmlDecoder.CharsetReader = charset.NewReader

	var rss struct {
		Channel Channel `xml:"channel"`
	}
	if err := xmlDecoder.Decode(&rss); err != nil {
		return nil, err
	}
	return &rss.Channel, nil
}
