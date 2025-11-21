package rss

import (
	"context"
	"encoding/xml"
	"io"
	"net/http"

	"github.com/paulrosania/go-charset/charset"
)

// Feed represents an Atom feed containing entries.
// It follows the Atom 1.0 specification structure.
type Feed struct {
	// Entry is a slice of entries in the feed
	Entry []Entry `xml:"entry"`
}

// Entry represents a single entry in an Atom feed.
// Each entry typically represents a blog post, article, or other piece of content.
type Entry struct {
	// ID is a permanent, universally unique identifier for the entry
	ID string `xml:"id"`

	// Title is the title of the entry
	Title string `xml:"title"`

	// Updated is the time when the entry was last modified
	Updated string `xml:"updated"`
}

// ParseAtom parses an Atom 1.0 feed from an io.Reader.
// It expects the reader to contain valid Atom XML.
// The context is used for cancellation control during parsing.
//
// The function automatically handles character encoding detection and conversion
// using the go-charset library, supporting various encodings commonly found
// in Atom feeds.
//
// Returns a Feed struct containing the parsed Atom data and any error that occurred.
// The reader is not closed by this function; the caller is responsible for closing it.
func ParseAtom(ctx context.Context, r io.Reader) (*Feed, error) {
	// Check if context is cancelled before starting
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	xmlDecoder := xml.NewDecoder(r)
	xmlDecoder.CharsetReader = charset.NewReader
	feed := Feed{}
	if err := xmlDecoder.Decode(&feed); err != nil {
		return nil, err
	}
	return &feed, nil
}

// Atom parses an Atom 1.0 feed from an HTTP response.
// It expects the response body to contain valid Atom XML.
// The context is used for cancellation control during parsing.
//
// The function automatically handles character encoding detection and conversion
// using the go-charset library, supporting various encodings commonly found
// in Atom feeds.
//
// Returns a Feed struct containing the parsed Atom data and any error that occurred.
// The response body is automatically closed after parsing.
func Atom(ctx context.Context, resp *http.Response) (*Feed, error) {
	defer resp.Body.Close()
	return ParseAtom(ctx, resp.Body)
}
