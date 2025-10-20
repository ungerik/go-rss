package rss

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"time"
)

const wordpressDateFormat = "Mon, 02 Jan 2006 15:04:05 -0700"

// Fetcher interface
type Fetcher interface {
	Get(url string) (resp *http.Response, err error)
}

// Date type
type Date string

// Parse (Date function) and returns Time, error
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

// ParseWithFormat (Date function), takes a string and returns Time, error
func (d Date) ParseWithFormat(format string) (time.Time, error) {
	return time.Parse(format, string(d))
}

// Format (Date function), takes a string and returns string, error
func (d Date) Format(format string) (string, error) {
	t, err := d.Parse()
	if err != nil {
		return "", err
	}
	return t.Format(format), nil
}

// MustFormat (Date function), take a string and returns string
func (d Date) MustFormat(format string) string {
	s, err := d.Format(format)
	if err != nil {
		return err.Error()
	}
	return s
}

// Read a string url and returns a Channel struct, error
func Read(url string, reddit bool) (*http.Response, error) {
	return ReadWithClient(url, http.DefaultClient, reddit)
}

// InsecureRead reads without certificate check
func InsecureRead(url string, reddit bool) (*http.Response, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	return ReadWithClient(url, client, reddit)
}

// ReadWithClient a string url and custom client that must match the Fetcher interface
// returns a Channel struct, error
func ReadWithClient(url string, client *http.Client, reddit bool) (*http.Response, error) {
	// Basic URL validation
	if url == "" {
		return nil, fmt.Errorf("URL cannot be empty")
	}

	req, err := http.NewRequest("GET", url, nil)
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
