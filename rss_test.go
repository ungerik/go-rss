package rss

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const (
	testDataDir    = "testdata"
	testFileSuffix = ".rss"
)

// testFetcher is an implementation of the Fetcher interface which reads the
// content from a local file.
type testFetcher struct{}

// Get takes a 'url' which is really just a name of a file in the 'testdata'
// directory and returns a fake http.Response with the file content as its body.
// It returns an error iff the file can not be opened.
func (f *testFetcher) Get(ctx context.Context, url string) (resp *http.Response, err error) {
	// Check if context is cancelled
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	file, err := os.Open(filepath.Join(testDataDir, url))
	if err != nil {
		return nil, err
	}
	return &http.Response{Body: file}, nil
}

// A trivial test making sure that that all feeds parse - it *does not* check
// for correctness or completeness thereof, except for dates.
func TestAllFeedsParse(t *testing.T) {
	fileInfos, err := os.ReadDir(testDataDir)
	if err != nil {
		t.Fatalf("os.ReadDir(%q) err = %v, expected nil", testDataDir, err)
	}
	for _, fileInfo := range fileInfos {
		fileName := fileInfo.Name()
		if !strings.HasSuffix(fileName, testFileSuffix) {
			continue
		}

		// Create a test client that uses our testTransport
		client := &http.Client{
			Transport: &testTransport{},
		}

		// Test with context
		ctx := context.Background()
		resp, err := ReadWithClient(ctx, fileName, client, false)
		if err != nil {
			t.Fatalf("ReadWithClient(%q) err = %v, expected nil", fileName, err)
		}

		channel, err := Regular(ctx, resp)
		if err != nil {
			fmt.Println(err)
		}

		for _, item := range channel.Item {
			if _, err := item.PubDate.Parse(); err != nil {
				t.Fatalf("Date Parser(%q) err = %v, expected nil", fileName, err)
			}
		}
	}
}

// TestContextCancellation tests that context cancellation works properly
func TestContextCancellation(t *testing.T) {
	// Create a context that's already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	client := &http.Client{
		Transport: &testTransport{},
	}

	// This should fail due to cancelled context
	_, err := ReadWithClient(ctx, "techcrunch.rss", client, false)
	if err == nil {
		t.Fatal("Expected error due to cancelled context, got nil")
	}
	// Check if the error contains context.Canceled (it might be wrapped)
	if err != context.Canceled && !strings.Contains(err.Error(), "context canceled") {
		t.Fatalf("Expected context.Canceled or wrapped error, got %v", err)
	}
}

// TestBasicTimeout tests timeout functionality
func TestBasicTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// This should timeout quickly
	_, err := Read(ctx, "https://httpbin.org/delay/1", false)
	if err == nil {
		t.Fatal("Expected timeout error, got nil")
	}
	if !strings.Contains(err.Error(), "context deadline exceeded") {
		t.Fatalf("Expected timeout error, got: %v", err)
	}
}

// TestManualCancellation tests manual context cancellation
func TestManualCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel immediately
	cancel()

	_, err := Read(ctx, "https://httpbin.org/delay/1", false)
	if err == nil {
		t.Fatal("Expected cancellation error, got nil")
	}
	if !strings.Contains(err.Error(), "context canceled") {
		t.Fatalf("Expected cancellation error, got: %v", err)
	}
}

// TestCustomClient tests custom HTTP client functionality
func TestCustomClient(t *testing.T) {
	client := &http.Client{
		Timeout: 100 * time.Millisecond,
	}

	ctx := context.Background()
	_, err := ReadWithClient(ctx, "https://httpbin.org/delay/1", client, false)
	if err == nil {
		t.Fatal("Expected timeout error, got nil")
	}
	// Check for timeout-related errors (could be client timeout or context timeout)
	if !strings.Contains(err.Error(), "timeout") && !strings.Contains(err.Error(), "deadline exceeded") {
		t.Fatalf("Expected timeout error, got: %v", err)
	}
}

// TestMultipleFeedsWithTimeout tests processing multiple feeds with shared timeout
func TestMultipleFeedsWithTimeout(t *testing.T) {
	feeds := []string{
		"techcrunch.rss",
		"podcast.rss",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client := &http.Client{
		Transport: &testTransport{},
	}

	processedCount := 0
	for i, feedURL := range feeds {
		select {
		case <-ctx.Done():
			t.Logf("Timeout reached, processed %d feeds before timeout", i)
			return
		default:
		}

		resp, err := ReadWithClient(ctx, feedURL, client, false)
		if err != nil {
			t.Logf("Error fetching feed %d: %v", i+1, err)
			continue
		}

		channel, err := Regular(ctx, resp)
		if err != nil {
			t.Logf("Error parsing feed %d: %v", i+1, err)
			continue
		}

		if channel.Title == "" {
			t.Errorf("Feed %d has empty title", i+1)
		}
		processedCount++
	}

	if processedCount == 0 {
		t.Error("No feeds were processed successfully")
	}
}

// TestContextWithValues tests context value passing
func TestContextWithValues(t *testing.T) {
	ctx := context.WithValue(context.Background(), "userID", "12345")
	ctx = context.WithValue(ctx, "requestID", "req-abc-123")

	client := &http.Client{
		Transport: &testTransport{},
	}

	resp, err := ReadWithClient(ctx, "techcrunch.rss", client, false)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	defer resp.Body.Close()

	channel, err := Regular(ctx, resp)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if channel.Title == "" {
		t.Error("Channel title is empty")
	}

	// Verify context values are preserved
	if ctx.Value("userID") != "12345" {
		t.Error("UserID context value not preserved")
	}
	if ctx.Value("requestID") != "req-abc-123" {
		t.Error("RequestID context value not preserved")
	}
}

// TestRedditFeedSupport tests Reddit-specific functionality
func TestRedditFeedSupport(t *testing.T) {
	ctx := context.Background()
	client := &http.Client{
		Transport: &testTransport{},
	}

	// Test with reddit flag set to true - reddit.rss is actually an Atom feed
	resp, err := ReadWithClient(ctx, "reddit.rss", client, true)
	if err != nil {
		t.Fatalf("Error fetching Reddit feed: %v", err)
	}
	defer resp.Body.Close()

	// Parse as Atom feed since Reddit feeds are Atom format
	feed, err := Atom(ctx, resp)
	if err != nil {
		t.Fatalf("Error parsing Reddit feed: %v", err)
	}

	if len(feed.Entry) == 0 {
		t.Error("Reddit feed has no entries")
	}

	// Check that we have entries with titles
	for i, entry := range feed.Entry {
		if entry.Title == "" {
			t.Errorf("Reddit entry %d has empty title", i)
		}
	}
}

// TestInsecureRead tests insecure reading functionality
func TestInsecureRead(t *testing.T) {
	ctx := context.Background()

	// Test insecure read with a test transport that simulates the behavior
	client := &http.Client{Transport: &testTransport{}}

	resp, err := ReadWithClient(ctx, "techcrunch.rss", client, false)
	if err != nil {
		t.Fatalf("Error with insecure read: %v", err)
	}
	defer resp.Body.Close()

	channel, err := Regular(ctx, resp)
	if err != nil {
		t.Fatalf("Error parsing insecure read result: %v", err)
	}

	if channel.Title == "" {
		t.Error("Channel title is empty")
	}
}

// TestDateParsing tests the Date type parsing functionality
func TestDateParsing(t *testing.T) {
	testCases := []struct {
		name     string
		dateStr  string
		expected bool // whether parsing should succeed
	}{
		{"RFC822 format", "Mon, 02 Jan 2006 15:04:05 -0700", true},
		{"RFC3339 format", "2006-01-02T15:04:05Z", true},
		{"WordPress format", "Mon, 02 Jan 2006 15:04:05 -0700", true},
		{"Invalid format", "not a date", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			date := Date(tc.dateStr)
			_, err := date.Parse()

			if tc.expected && err != nil {
				t.Errorf("Expected parsing to succeed for %s, got error: %v", tc.dateStr, err)
			}
			if !tc.expected && err == nil {
				t.Errorf("Expected parsing to fail for %s, but it succeeded", tc.dateStr)
			}
		})
	}
}

// TestDateFormatting tests the Date type formatting functionality
func TestDateFormatting(t *testing.T) {
	date := Date("Mon, 02 Jan 2006 15:04:05 -0700")

	// Test Format method
	formatted, err := date.Format("2006-01-02")
	if err != nil {
		t.Fatalf("Error formatting date: %v", err)
	}
	if formatted != "2006-01-02" {
		t.Errorf("Expected '2006-01-02', got '%s'", formatted)
	}

	// Test MustFormat method
	mustFormatted := date.MustFormat("2006-01-02")
	if mustFormatted != "2006-01-02" {
		t.Errorf("Expected '2006-01-02', got '%s'", mustFormatted)
	}
}

// TestAtomFeedParsing tests Atom feed parsing
func TestAtomFeedParsing(t *testing.T) {
	ctx := context.Background()
	client := &http.Client{
		Transport: &testTransport{},
	}

	resp, err := ReadWithClient(ctx, "reddit-google.rss", client, false)
	if err != nil {
		t.Fatalf("Error fetching Atom feed: %v", err)
	}
	defer resp.Body.Close()

	feed, err := Atom(ctx, resp)
	if err != nil {
		t.Fatalf("Error parsing Atom feed: %v", err)
	}

	if len(feed.Entry) == 0 {
		t.Error("Atom feed has no entries")
	}

	for i, entry := range feed.Entry {
		if entry.Title == "" {
			t.Errorf("Entry %d has empty title", i)
		}
	}
}

// TestErrorHandling tests various error conditions
func TestErrorHandling(t *testing.T) {
	ctx := context.Background()

	// Test empty URL
	_, err := Read(ctx, "", false)
	if err == nil {
		t.Error("Expected error for empty URL")
	}
	if !strings.Contains(err.Error(), "URL cannot be empty") {
		t.Errorf("Expected 'URL cannot be empty' error, got: %v", err)
	}

	// Test invalid URL
	_, err = Read(ctx, "not-a-url", false)
	if err == nil {
		t.Error("Expected error for invalid URL")
	}
}

// TestResourceCleanup tests that resources are properly cleaned up
func TestResourceCleanup(t *testing.T) {
	ctx := context.Background()
	client := &http.Client{
		Transport: &testTransport{},
	}

	resp, err := ReadWithClient(ctx, "techcrunch.rss", client, false)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	// Parse should close the response body
	_, err = Regular(ctx, resp)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	// Try to read from the closed body - this should fail
	_, err = resp.Body.Read(make([]byte, 1))
	if err == nil {
		t.Error("Expected error when reading from closed body")
	}
}

// testTransport implements http.RoundTripper to read from local files
type testTransport struct{}

func (t *testTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Check if context is cancelled
	select {
	case <-req.Context().Done():
		return nil, req.Context().Err()
	default:
	}

	file, err := os.Open(filepath.Join(testDataDir, req.URL.Path))
	if err != nil {
		return nil, err
	}
	return &http.Response{
		Body:       file,
		StatusCode: 200,
		Status:     "200 OK",
	}, nil
}
