//Package rss provides a Simple RSS parser, tested with various feeds.
package rss

import (
	"encoding/xml"
	"net/http"
	"time"
	"crypto/tls"

	"github.com/paulrosania/go-charset/charset"
	_ "github.com/paulrosania/go-charset/data" //initialize only
)

const (
	wordpressDateFormat = "Mon, 02 Jan 2006 15:04:05 -0700"
)

//Fetcher interface
type Fetcher interface {
	Get(url string) (resp *http.Response, err error)
}

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
	Author 		string          `xml:"author"`
	Content     string          `xml:"content"`
	FullText    string          `xml:"full-text"`
}

//Date type
type Date string

//Parse (Date function) and returns Time, error
func (d Date) Parse() (time.Time, error) {
	t, err := d.ParseWithFormat(wordpressDateFormat)
	if err != nil {
		t, err = d.ParseWithFormat(time.RFC822) // RSS 2.0 spec
	}
	return t, err
}

//ParseWithFormat (Date function), takes a string and returns Time, error
func (d Date) ParseWithFormat(format string) (time.Time, error) {
	return time.Parse(format, string(d))
}

//Format (Date function), takes a string and returns string, error
func (d Date) Format(format string) (string, error) {
	t, err := d.Parse()
	if err != nil {
		return "", err
	}
	return t.Format(format), nil
}

//MustFormat (Date function), take a string and returns string
func (d Date) MustFormat(format string) string {
	s, err := d.Format(format)
	if err != nil {
		return err.Error()
	}
	return s
}

//Read a string url and returns a Channel struct, error
func Read(url string) (*Channel, error) {
	return ReadWithClient(url, http.DefaultClient)
}

//Read without certificate check
func InsecureRead(url string) (*Channel, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	return ReadWithClient(url, client)
}

//ReadWithClient a string url and custom client that must match the Fetcher interface
//returns a Channel struct, error
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
