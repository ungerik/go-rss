package rss

import (
	"io/ioutil"
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
// for correctness or completeness thereof.
func TestAllFeedsParse(t *testing.T) {
	fileInfos, err := ioutil.ReadDir(testDataDir)
	if err != nil {
		t.Fatalf("ioutil.ReadDir(%q) err = %v, expected nil", testDataDir, err)
	}
	for _, fileInfo := range fileInfos {
		fileName := fileInfo.Name()
		if !strings.HasSuffix(fileName, testFileSuffix) {
			continue
		}
		if _, err := ReadWithClient(fileName, new(testFetcher)); err != nil {
			t.Fatalf("ReadWithClient(%q) err = %v, expected nil", fileName, err)
		}
	}
}
