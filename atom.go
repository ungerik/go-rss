package rss

import (
	"encoding/xml"
	"net/http"

	"github.com/paulrosania/go-charset/charset"
)

//Feed struct for RSS
type Feed struct {
	Entry []Entry `xml:"entry"`
}

//Entry struct for each Entry in the Feed
type Entry struct {
	ID      string `xml:"id"`
	Title   string `xml:"title"`
	Updated string `xml:"updated"`
}

//Atom parses atom feeds
func Atom(resp *http.Response) (*Feed, error) {
	defer resp.Body.Close()
	xmlDecoder := xml.NewDecoder(resp.Body)
	xmlDecoder.CharsetReader = charset.NewReader
	feed := Feed{}
	if err := xmlDecoder.Decode(&feed); err != nil {
		return nil, err
	}
	return &feed, nil
}
