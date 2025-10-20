// Package rss provides a simple RSS and Atom feed parser with context support.
// It supports RSS 2.0 and Atom feeds with proper error handling, timeout support,
// and resource management.
//
// Basic usage:
//
//	ctx := context.Background()
//	resp, err := rss.Read(ctx, "https://example.com/feed.rss", false)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer resp.Body.Close()
//
//	channel, err := rss.Regular(ctx, resp)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	fmt.Printf("Channel: %s\n", channel.Title)
package rss

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"time"
)

// wordpressDateFormat is the date format commonly used by WordPress RSS feeds.
const wordpressDateFormat = "Mon, 02 Jan 2006 15:04:05 -0700"

// Fetcher defines the interface for fetching HTTP resources.
// This interface allows for custom implementations of HTTP clients
// while maintaining compatibility with the RSS parsing functions.
type Fetcher interface {
	// Get fetches a resource from the given URL using the provided context.
	// It returns an HTTP response and any error that occurred during the fetch.
	Get(ctx context.Context, url string) (resp *http.Response, err error)
}

// Date represents a date string from RSS/Atom feeds.
// It provides methods for parsing dates in various formats commonly found
// in RSS and Atom feeds.
type Date string

// Parse attempts to parse the date string using multiple common formats.
// It tries formats in the following order:
// 1. WordPress format (Mon, 02 Jan 2006 15:04:05 -0700)
// 2. RFC822 format (RSS 2.0 standard)
// 3. RFC3339 format (Atom standard)
//
// Returns the parsed time and any error that occurred.
func (d Date) Parse() (time.Time, error) {
	t, err := d.ParseWithFormat(wordpressDateFormat)
	if err != nil {
		t, err = d.ParseWithFormat(time.RFC822) // RSS 2.0 spec
		if err != nil {
			t, err = d.ParseWithFormat(time.RFC3339) // Atom
		}
	}
	return t, err
}

// ParseWithFormat parses the date string using the specified format.
// The format should follow Go's time format layout (reference time: Mon Jan 2 15:04:05 MST 2006).
//
// Returns the parsed time and any error that occurred.
func (d Date) ParseWithFormat(format string) (time.Time, error) {
	return time.Parse(format, string(d))
}

// Format parses the date and formats it using the specified format string.
// It first calls Parse() to convert the date string to a time.Time,
// then formats it using the provided format.
//
// Returns the formatted date string and any error that occurred.
func (d Date) Format(format string) (string, error) {
	t, err := d.Parse()
	if err != nil {
		return "", err
	}
	return t.Format(format), nil
}

// MustFormat parses the date and formats it using the specified format string.
// Unlike Format(), this function does not return an error. If parsing fails,
// it returns the error message as a string instead of panicking.
//
// This is useful when you want to display a date but don't want to handle
// parsing errors explicitly.
func (d Date) MustFormat(format string) string {
	s, err := d.Format(format)
	if err != nil {
		return err.Error()
	}
	return s
}

// Read fetches an RSS or Atom feed from the given URL using the default HTTP client.
// The context is used for cancellation and timeout control.
// The reddit parameter should be set to true when fetching Reddit feeds to use
// the appropriate user agent header.
//
// Returns an HTTP response that should be closed by the caller.
// The response body should be passed to either Regular() or Atom() for parsing.
func Read(ctx context.Context, url string, reddit bool) (*http.Response, error) {
	return ReadWithClient(ctx, url, http.DefaultClient, reddit)
}

// InsecureRead fetches an RSS or Atom feed from the given URL using an HTTP client
// that skips SSL certificate verification. This should only be used in development
// or testing environments where SSL verification is not critical.
//
// WARNING: This function disables SSL certificate verification, making it vulnerable
// to man-in-the-middle attacks. Do not use in production environments.
func InsecureRead(ctx context.Context, url string, reddit bool) (*http.Response, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	return ReadWithClient(ctx, url, client, reddit)
}

// ReadWithClient fetches an RSS or Atom feed from the given URL using a custom HTTP client.
// This allows for custom transport configurations, timeouts, and other HTTP client settings.
// The context is used for cancellation and timeout control.
//
// The client parameter allows you to customize the HTTP behavior, such as:
// - Setting custom timeouts
// - Configuring proxy settings
// - Adding custom headers
// - Implementing custom transport logic
//
// Returns an HTTP response that should be closed by the caller.
// The response body should be passed to either Regular() or Atom() for parsing.
func ReadWithClient(ctx context.Context, url string, client *http.Client, reddit bool) (*http.Response, error) {
	// Basic URL validation
	if url == "" {
		return nil, fmt.Errorf("URL cannot be empty")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set appropriate user agent
	if reddit {
		// This header is required to read Reddit Feeds, see:
		// https://www.reddit.com/r/redditdev/comments/5w60r1/error_429_too_many_requests_i_havent_made_many/
		// Note: a random string is required to prevent occurrence of 'Too Many Requests' response.
		req.Header.Set("user-agent", "go-rss:v1.0.0 (by /u/go-rss-user)")
	} else {
		// Set a generic user agent for other feeds
		req.Header.Set("user-agent", "go-rss/1.0.0")
	}

	response, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}

	// Check for successful response
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		response.Body.Close()
		return nil, fmt.Errorf("HTTP %d: %s", response.StatusCode, response.Status)
	}

	return response, nil
}
