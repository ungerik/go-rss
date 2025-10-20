package rss

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
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
func (f *testFetcher) Get(url string) (resp *http.Response, err error) {
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

		// Create a test client that uses our testFetcher
		client := &http.Client{
			Transport: &testTransport{},
		}

		resp, err := ReadWithClient(fileName, client, false)
		if err != nil {
			t.Fatalf("ReadWithClient(%q) err = %v, expected nil", fileName, err)
		}

		channel, err := Regular(resp)
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

// testTransport implements http.RoundTripper to read from local files
type testTransport struct{}

func (t *testTransport) RoundTrip(req *http.Request) (*http.Response, error) {
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
