# go-rss

Simple RSS parser with context support for cancellation and timeouts, tested with various feeds.

## Features

- RSS 2.0 and Atom feed parsing
- Context support for cancellation and timeouts
- Custom HTTP client support
- Reddit feed support with proper user agent
- Comprehensive error handling
- Resource leak prevention

## Usage

### Basic Usage

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
    
    resp, err := rss.Read(ctx, "https://example.com/feed.rss", false)
    if err != nil {
        log.Fatal(err)
    }
    defer resp.Body.Close()

    channel, err := rss.Regular(ctx, resp)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Channel: %s\n", channel.Title)
    for _, item := range channel.Item {
        fmt.Printf("- %s\n", item.Title)
    }
}
```

### Context Support

```go
package main

import (
    "context"
    "fmt"
    "time"
    "github.com/ungerik/go-rss"
)

func main() {
    // Create context with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // Use context-aware functions
    resp, err := rss.Read(ctx, "https://example.com/feed.rss", false)
    if err != nil {
        log.Fatal(err)
    }
    defer resp.Body.Close()

    channel, err := rss.Regular(ctx, resp)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Channel: %s\n", channel.Title)
}
```

### Available Functions

#### Reading Feeds
- `Read(ctx context.Context, url string, reddit bool) (*http.Response, error)` - Basic feed reading
- `ReadWithClient(ctx context.Context, url string, client *http.Client, reddit bool) (*http.Response, error)` - With custom client
- `InsecureRead(ctx context.Context, url string, reddit bool) (*http.Response, error)` - Without SSL verification

#### Parsing Feeds
- `Regular(ctx context.Context, resp *http.Response) (*Channel, error)` - Parse RSS feeds
- `Atom(ctx context.Context, resp *http.Response) (*Feed, error)` - Parse Atom feeds

## Examples

See the `example` folder for:
- `main.go` - Basic usage with context and timeout
- `context_examples.go` - Comprehensive context usage examples

## Testing

Run the tests with:
```bash
go test -v
```

The tests include:
- Feed parsing validation
- Context cancellation testing
- Context timeout testing