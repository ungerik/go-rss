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

// TestContextTimeout tests that context timeout works properly
func TestContextTimeout(t *testing.T) {
	// Create a context with a very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Wait a bit to ensure timeout
	time.Sleep(1 * time.Millisecond)

	client := &http.Client{
		Transport: &testTransport{},
	}

	// This should fail due to timeout
	_, err := ReadWithClient(ctx, "techcrunch.rss", client, false)
	if err == nil {
		t.Fatal("Expected error due to timeout, got nil")
	}
	// Check if the error contains context.DeadlineExceeded (it might be wrapped)
	if err != context.DeadlineExceeded && !strings.Contains(err.Error(), "context deadline exceeded") {
		t.Fatalf("Expected context.DeadlineExceeded or wrapped error, got %v", err)
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
