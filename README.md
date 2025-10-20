# go-rss

[![Go Reference](https://pkg.go.dev/badge/github.com/ungerik/go-rss.svg)](https://pkg.go.dev/github.com/ungerik/go-rss)
[![Go Report Card](https://goreportcard.com/badge/github.com/ungerik/go-rss)](https://goreportcard.com/report/github.com/ungerik/go-rss)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A simple, robust RSS and Atom feed parser for Go with comprehensive context support for cancellation and timeouts. This library provides a clean API for fetching and parsing RSS 2.0 and Atom 1.0 feeds with proper error handling, resource management, and modern Go conventions.

## Features

- **RSS 2.0 and Atom 1.0 Support** - Parse both major feed formats
- **Context Support** - Full cancellation and timeout support using `context.Context`
- **Custom HTTP Clients** - Use your own HTTP client configurations
- **Reddit Feed Support** - Special handling for Reddit feeds with proper user agents
- **Character Encoding** - Automatic detection and conversion of various encodings
- **Resource Management** - Proper cleanup of HTTP response bodies
- **Comprehensive Error Handling** - Detailed error messages with context
- **Modern Go Conventions** - Context as first parameter, error wrapping
- **Zero Dependencies** - Only uses standard library and one charset library

## Installation

```bash
go get github.com/ungerik/go-rss
```

## Quick Start

### Basic RSS Feed Reading

```go
package main

import (
    "context"
    "fmt"
    "log"
    "github.com/ungerik/go-rss"
)

func main() {
    ctx := context.Background()
    
    // Fetch an RSS feed
    resp, err := rss.Read(ctx, "https://example.com/feed.rss", false)
    if err != nil {
        log.Fatal(err)
    }
    defer resp.Body.Close()

    // Parse the RSS feed
    channel, err := rss.Regular(ctx, resp)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Channel: %s\n", channel.Title)
    fmt.Printf("Description: %s\n", channel.Description)
    
    for _, item := range channel.Item {
        fmt.Printf("- %s (%s)\n", item.Title, item.Link)
    }
}
```

### Atom Feed Reading

```go
package main

import (
    "context"
    "fmt"
    "log"
    "github.com/ungerik/go-rss"
)

func main() {
    ctx := context.Background()
    
    // Fetch an Atom feed
    resp, err := rss.Read(ctx, "https://example.com/feed.atom", false)
    if err != nil {
        log.Fatal(err)
    }
    defer resp.Body.Close()

    // Parse the Atom feed
    feed, err := rss.Atom(ctx, resp)
    if err != nil {
        log.Fatal(err)
    }

    for _, entry := range feed.Entry {
        fmt.Printf("- %s (Updated: %s)\n", entry.Title, entry.Updated)
    }
}
```

## API Reference

### Reading Feeds

#### `Read(ctx context.Context, url string, reddit bool) (*http.Response, error)`

Fetches an RSS or Atom feed from the given URL using the default HTTP client.

**Parameters:**
- `ctx` - Context for cancellation and timeout control
- `url` - The URL of the feed to fetch
- `reddit` - Set to `true` for Reddit feeds to use appropriate user agent

**Returns:**
- `*http.Response` - HTTP response (caller must close the body)
- `error` - Any error that occurred during fetching

#### `ReadWithClient(ctx context.Context, url string, client *http.Client, reddit bool) (*http.Response, error)`

Fetches a feed using a custom HTTP client, allowing for custom configurations.

**Use cases:**
- Custom timeouts
- Proxy settings
- Custom headers
- Custom transport logic

#### `InsecureRead(ctx context.Context, url string, reddit bool) (*http.Response, error)`

Fetches a feed without SSL certificate verification.

**⚠️ Warning:** This disables SSL verification and should only be used in development or testing environments.

### Parsing Feeds

#### `Regular(ctx context.Context, resp *http.Response) (*Channel, error)`

Parses an RSS 2.0 feed from an HTTP response.

**Returns:**
- `*Channel` - Parsed RSS channel data
- `error` - Any parsing error

#### `Atom(ctx context.Context, resp *http.Response) (*Feed, error)`

Parses an Atom 1.0 feed from an HTTP response.

**Returns:**
- `*Feed` - Parsed Atom feed data
- `error` - Any parsing error

### Data Structures

#### Channel (RSS)

```go
type Channel struct {
    Title         string `xml:"title"`         // Channel title
    Link          string `xml:"link"`          // Channel URL
    Description   string `xml:"description"`   // Channel description
    Language      string `xml:"language"`      // Channel language
    LastBuildDate Date   `xml:"lastBuildDate"` // Last build date
    Item          []Item `xml:"item"`          // Channel items
}
```

#### Item (RSS)

```go
type Item struct {
    Title       string          `xml:"title"`       // Item title
    Link        string          `xml:"link"`        // Item URL
    Comments    string          `xml:"comments"`   // Comments URL
    PubDate     Date            `xml:"pubDate"`     // Publication date
    GUID        string          `xml:"guid"`        // Unique identifier
    Category    []string        `xml:"category"`   // Categories
    Enclosure   []ItemEnclosure `xml:"enclosure"`  // Media enclosures
    Description string          `xml:"description"` // Item description
    Author      string          `xml:"author"`     // Author email
    Content     string          `xml:"content"`    // Full content
    FullText    string          `xml:"full-text"`  // Complete text
}
```

#### Feed (Atom)

```go
type Feed struct {
    Entry []Entry `xml:"entry"` // Feed entries
}

type Entry struct {
    ID      string `xml:"id"`      // Unique identifier
    Title   string `xml:"title"`  // Entry title
    Updated string `xml:"updated"` // Last updated time
}
```

### Date Handling

The `Date` type provides methods for parsing dates in various formats:

```go
// Parse using multiple common formats
t, err := item.PubDate.Parse()

// Parse with specific format
t, err := item.PubDate.ParseWithFormat(time.RFC822)

// Format parsed date
formatted, err := item.PubDate.Format("2006-01-02 15:04:05")

// Format without error handling (returns error string on failure)
formatted := item.PubDate.MustFormat("2006-01-02")
```

## Advanced Usage

### Context with Timeout

```go
package main

import (
    "context"
    "fmt"
    "time"
    "github.com/ungerik/go-rss"
)

func main() {
    // Create context with 10 second timeout
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    resp, err := rss.Read(ctx, "https://slow-feed.com/rss", false)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    defer resp.Body.Close()

    channel, err := rss.Regular(ctx, resp)
    if err != nil {
        fmt.Printf("Parse error: %v\n", err)
        return
    }

    fmt.Printf("Successfully parsed: %s\n", channel.Title)
}
```

### Context with Cancellation

```go
package main

import (
    "context"
    "fmt"
    "time"
    "github.com/ungerik/go-rss"
)

func main() {
    ctx, cancel := context.WithCancel(context.Background())
    
    // Cancel after 5 seconds
    go func() {
        time.Sleep(5 * time.Second)
        cancel()
        fmt.Println("Operation cancelled")
    }()

    resp, err := rss.Read(ctx, "https://example.com/feed.rss", false)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    defer resp.Body.Close()

    // ... rest of processing
}
```

### Custom HTTP Client

```go
package main

import (
    "context"
    "fmt"
    "net/http"
    "time"
    "github.com/ungerik/go-rss"
)

func main() {
    // Create custom HTTP client with timeout
    client := &http.Client{
        Timeout: 30 * time.Second,
        Transport: &http.Transport{
            MaxIdleConns:        10,
            IdleConnTimeout:     30 * time.Second,
            DisableCompression:  true,
        },
    }

    ctx := context.Background()
    resp, err := rss.ReadWithClient(ctx, "https://example.com/feed.rss", client, false)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    defer resp.Body.Close()

    channel, err := rss.Regular(ctx, resp)
    if err != nil {
        fmt.Printf("Parse error: %v\n", err)
        return
    }

    fmt.Printf("Channel: %s\n", channel.Title)
}
```

### Reddit Feed Support

```go
package main

import (
    "context"
    "fmt"
    "github.com/ungerik/go-rss"
)

func main() {
    ctx := context.Background()
    
    // Fetch Reddit feed with special user agent
    resp, err := rss.Read(ctx, "https://reddit.com/r/golang.rss", true)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    defer resp.Body.Close()

    channel, err := rss.Regular(ctx, resp)
    if err != nil {
        fmt.Printf("Parse error: %v\n", err)
        return
    }

    fmt.Printf("Reddit Channel: %s\n", channel.Title)
}
```

### Processing Multiple Feeds

```go
package main

import (
    "context"
    "fmt"
    "strings"
    "sync"
    "time"
    "github.com/ungerik/go-rss"
)

func main() {
    feeds := []string{
        "https://blog.golang.org/feed.atom",
        "https://example.com/feed.rss",
        "https://reddit.com/r/golang.rss",
    }

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    var wg sync.WaitGroup
    results := make(chan string, len(feeds))

    for i, feedURL := range feeds {
        wg.Add(1)
        go func(index int, url string) {
            defer wg.Done()
            
            isReddit := false
            if strings.Contains(url, "reddit.com") {
                isReddit = true
            }

            resp, err := rss.Read(ctx, url, isReddit)
            if err != nil {
                results <- fmt.Sprintf("Feed %d error: %v", index, err)
                return
            }
            defer resp.Body.Close()

            // Determine feed type and parse accordingly
            if strings.Contains(url, ".atom") || isReddit {
                feed, err := rss.Atom(ctx, resp)
                if err != nil {
                    results <- fmt.Sprintf("Feed %d parse error: %v", index, err)
                    return
                }
                results <- fmt.Sprintf("Feed %d: %d entries", index, len(feed.Entry))
            } else {
                channel, err := rss.Regular(ctx, resp)
                if err != nil {
                    results <- fmt.Sprintf("Feed %d parse error: %v", index, err)
                    return
                }
                results <- fmt.Sprintf("Feed %d: %s (%d items)", index, channel.Title, len(channel.Item))
            }
        }(i, feedURL)
    }

    go func() {
        wg.Wait()
        close(results)
    }()

    for result := range results {
        fmt.Println(result)
    }
}
```

## Error Handling

The library provides comprehensive error handling with wrapped errors for better debugging:

```go
resp, err := rss.Read(ctx, "https://example.com/feed.rss", false)
if err != nil {
    // Check for specific error types
    if strings.Contains(err.Error(), "context deadline exceeded") {
        fmt.Println("Request timed out")
    } else if strings.Contains(err.Error(), "context canceled") {
        fmt.Println("Request was cancelled")
    } else {
        fmt.Printf("Other error: %v\n", err)
    }
    return
}
```

## Testing

Run the test suite:

```bash
go test -v
```

The tests include:
- Feed parsing validation
- Context cancellation testing
- Context timeout testing
- Error handling validation

## Examples

See the `example/` and `examples/` directories for comprehensive usage examples:

- `example/main.go` - Basic usage with context and timeout
- `examples/context/main.go` - Advanced context usage patterns

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Changelog

### v2.0.0 (Breaking Changes)
- Added `context.Context` as first parameter to all functions
- Removed `WithContext` suffix variants
- Improved error handling with wrapped errors
- Added comprehensive documentation
- Enhanced resource management

### v1.0.0
- Initial release with basic RSS/Atom parsing
- Context support with `WithContext` variants
- Reddit feed support