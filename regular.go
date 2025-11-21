package rss

import (
	"context"
	"encoding/xml"
	"io"
	"net/http"

	"github.com/paulrosania/go-charset/charset"
)

// Channel represents an RSS channel containing metadata and items.
// It follows the RSS 2.0 specification structure.
type Channel struct {
	// Title is the name of the channel
	Title string `xml:"title"`

	// Link is the URL to the HTML website corresponding to the channel
	Link string `xml:"link"`

	// Description is a phrase or sentence describing the channel
	Description string `xml:"description"`

	// Language is the language the channel is written in
	Language string `xml:"language"`

	// LastBuildDate indicates the last time the content of the channel changed
	LastBuildDate Date `xml:"lastBuildDate"`

	// Item is a slice of items in the channel
	Item []Item `xml:"item"`
}

// ItemEnclosure represents an enclosure element in an RSS item.
// Enclosures are used to include media files with RSS items.
type ItemEnclosure struct {
	// URL is the location of the enclosed file
	URL string `xml:"url,attr"`

	// Type is the MIME type of the enclosed file
	Type string `xml:"type,attr"`
}

// Item represents a single item in an RSS channel.
// Each item typically represents a story, article, or other piece of content.
type Item struct {
	// Title is the title of the item
	Title string `xml:"title"`

	// Link is the URL of the item
	Link string `xml:"link"`

	// Comments is the URL of a page for comments relating to the item
	Comments string `xml:"comments"`

	// PubDate is the publication date of the item
	PubDate Date `xml:"pubDate"`

	// GUID is a string that uniquely identifies the item
	GUID string `xml:"guid"`

	// Category is a list of categories that the item belongs to
	Category []string `xml:"category"`

	// Enclosure is a list of media files associated with the item
	Enclosure []ItemEnclosure `xml:"enclosure"`

	// Description is a synopsis of the item
	Description string `xml:"description"`

	// Author is the email address of the author of the item
	Author string `xml:"author"`

	// Content is the full content of the item (if available)
	Content string `xml:"content"`

	// FullText is the complete text content of the item
	FullText string `xml:"full-text"`
}

// ParseRegular parses an RSS 2.0 feed from an io.Reader.
// It expects the reader to contain valid RSS XML.
// The context is used for cancellation control during parsing.
//
// The function automatically handles character encoding detection and conversion
// using the go-charset library, supporting various encodings commonly found
// in RSS feeds.
//
// Returns a Channel struct containing the parsed RSS data and any error that occurred.
// The reader is not closed by this function; the caller is responsible for closing it.
func ParseRegular(ctx context.Context, r io.Reader) (*Channel, error) {
	// Check if context is cancelled before starting
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	xmlDecoder := xml.NewDecoder(r)
	xmlDecoder.CharsetReader = charset.NewReader

	var rss struct {
		Channel Channel `xml:"channel"`
	}
	if err := xmlDecoder.Decode(&rss); err != nil {
		return nil, err
	}
	return &rss.Channel, nil
}

// Regular parses an RSS 2.0 feed from an HTTP response.
// It expects the response body to contain valid RSS XML.
// The context is used for cancellation control during parsing.
//
// The function automatically handles character encoding detection and conversion
// using the go-charset library, supporting various encodings commonly found
// in RSS feeds.
//
// Returns a Channel struct containing the parsed RSS data and any error that occurred.
// The response body is automatically closed after parsing.
func Regular(ctx context.Context, resp *http.Response) (*Channel, error) {
	defer resp.Body.Close()
	return ParseRegular(ctx, resp.Body)
}
