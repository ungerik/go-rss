/*
Simple RSS parser, tested with Wordpress feeds.
*/
package rss

import (
	"encoding/xml"
	"net/http"
	"time"
)

type Feed struct {
	Title       string
	Link        string
	Description string
	Language    string
	Item        []Item
}

type Item struct {
	Title       string
	Link        string
	PubDate     string
	Description string
	Enclosure   struct {
		URL string `xml:"attr"`
	}
}

func (self *Item) ParsePubDate() (time.Time, error) {
	t, err := time.Parse("Mon, 02 Jan 2006 15:04:05 -0700", self.PubDate) // Wordpress format
	if err != nil {
		t, err = time.Parse(time.RFC822, self.PubDate) // RSS 2.0 spec
	}
	return t, err
}

func (self *Item) FormatPubDate(format string) (string, error) {
	t, err := self.ParsePubDate()
	if err != nil {
		return "", err
	}
	return t.Format(format), nil
}

func (self *Item) MustFormatPubDate(format string) string {
	s, err := self.FormatPubDate(format)
	if err != nil {
		return err.Error()
	}
	return s
}

func Read(url string) (feed *Feed, err error) {
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	var rss struct {
		Channel Feed
	}
	err = xml.Unmarshal(response.Body, &rss)
	if err != nil {
		return nil, err
	}
	return &rss.Channel, nil
}
