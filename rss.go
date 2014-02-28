/**
 * Simple RSS parser, tested with various feeds.
 */
package rss

import (
	"encoding/xml"
	"net/http"
	"time"

	"code.google.com/p/go-charset/charset"
	_ "code.google.com/p/go-charset/data"
)

const (
	wordpressDateFormat = "Mon, 02 Jan 2006 15:04:05 -0700"
)

type Fetcher interface {
	Get(url string) (resp *http.Response, err error)
}

type Channel struct {
	Title         string `xml:"title"`
	Link          string `xml:"link"`
	Description   string `xml:"description"`
	Language      string `xml:"language"`
	LastBuildDate Date   `xml:"lastBuildDate"`
	Item          []Item `xml:"item"`
}

type ItemEnclosure struct {
	URL  string `xml:"url,attr"`
	Type string `xml:"type,attr"`
}

type Item struct {
	Title       string        `xml:"title"`
	Link        string        `xml:"link"`
	Comments    string        `xml:"comments"`
	PubDate     Date          `xml:"pubDate"`
	GUID        string        `xml:"guid"`
	Category    []string      `xml:"category"`
	Enclosure   ItemEnclosure `xml:"enclosure"`
	Description string        `xml:"description"`
	Content     string        `xml:"content"`
}

type Date string

func (self Date) Parse() (time.Time, error) {
	t, err := self.ParseWithFormat(wordpressDateFormat)
	if err != nil {
		t, err = self.ParseWithFormat(time.RFC822) // RSS 2.0 spec
	}
	return t, err
}

func (self Date) ParseWithFormat(format string) (time.Time, error) {
	return time.Parse(format, string(self))
}

func (self Date) Format(format string) (string, error) {
	t, err := self.Parse()
	if err != nil {
		return "", err
	}
	return t.Format(format), nil
}

func (self Date) MustFormat(format string) string {
	s, err := self.Format(format)
	if err != nil {
		return err.Error()
	}
	return s
}

func Read(url string) (*Channel, error) {
	return ReadWithClient(url, http.DefaultClient)
}

func ReadWithClient(url string, client Fetcher) (*Channel, error) {
	response, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	xmlDecoder := xml.NewDecoder(response.Body)
	xmlDecoder.CharsetReader = charset.NewReader

	var rss struct {
		Channel Channel `xml:"channel"`
	}
	if err = xmlDecoder.Decode(&rss); err != nil {
		return nil, err
	}
	return &rss.Channel, nil
}
