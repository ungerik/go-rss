package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/ungerik/go-rss"
)

func main() {
	fmt.Println("=== Context Examples for go-rss ===")

	// Example 1: Basic timeout
	exampleTimeout()

	// Example 2: Cancellation
	exampleCancellation()

	// Example 3: Custom client with timeout
	exampleCustomClient()

	// Example 4: Multiple feeds with shared timeout
	exampleMultipleFeeds()

	// Example 5: Context with values
	exampleContextWithValues()
}

// Example 1: Basic timeout usage
func exampleTimeout() {
	fmt.Println("1. Basic Timeout Example:")
	fmt.Println("   Fetching feed with 5 second timeout...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Try to fetch a feed (this will likely timeout if the URL is slow)
	resp, err := rss.Read(ctx, "https://httpbin.org/delay/10", false)
	if err != nil {
		fmt.Printf("   Expected timeout error: %v\n", err)
	} else {
		resp.Body.Close()
		fmt.Println("   Unexpected success!")
	}
	fmt.Println()
}

// Example 2: Manual cancellation
func exampleCancellation() {
	fmt.Println("2. Manual Cancellation Example:")
	fmt.Println("   Starting fetch, will cancel after 1 second...")

	ctx, cancel := context.WithCancel(context.Background())

	// Cancel after 1 second
	go func() {
		time.Sleep(1 * time.Second)
		cancel()
		fmt.Println("   Cancelled!")
	}()

	resp, err := rss.Read(ctx, "https://httpbin.org/delay/5", false)
	if err != nil {
		fmt.Printf("   Expected cancellation error: %v\n", err)
	} else {
		resp.Body.Close()
		fmt.Println("   Unexpected success!")
	}
	fmt.Println()
}

// Example 3: Custom HTTP client with timeout
func exampleCustomClient() {
	fmt.Println("3. Custom HTTP Client Example:")
	fmt.Println("   Using custom client with 3 second timeout...")

	// Create a custom HTTP client with timeout
	client := &http.Client{
		Timeout: 3 * time.Second,
	}

	ctx := context.Background()
	resp, err := rss.ReadWithClient(ctx, "https://httpbin.org/delay/5", client, false)
	if err != nil {
		fmt.Printf("   Expected timeout error: %v\n", err)
	} else {
		resp.Body.Close()
		fmt.Println("   Unexpected success!")
	}
	fmt.Println()
}

// Example 4: Multiple feeds with shared timeout
func exampleMultipleFeeds() {
	fmt.Println("4. Multiple Feeds with Shared Timeout:")
	fmt.Println("   Processing multiple feeds with 10 second total timeout...")

	feeds := []string{
		"https://httpbin.org/delay/2",
		"https://httpbin.org/delay/3",
		"https://httpbin.org/delay/4",
		"https://httpbin.org/delay/5",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for i, feedURL := range feeds {
		select {
		case <-ctx.Done():
			fmt.Printf("   Timeout reached, processed %d feeds before timeout\n", i)
			return
		default:
		}

		fmt.Printf("   Processing feed %d: %s\n", i+1, feedURL)
		resp, err := rss.Read(ctx, feedURL, false)
		if err != nil {
			fmt.Printf("   Error fetching feed %d: %v\n", i+1, err)
			continue
		}

		// Parse the feed
		channel, err := rss.Regular(ctx, resp)
		if err != nil {
			fmt.Printf("   Error parsing feed %d: %v\n", i+1, err)
			continue
		}

		fmt.Printf("   Successfully parsed feed %d: %s\n", i+1, channel.Title)
	}

	fmt.Println("   All feeds processed successfully!")
	fmt.Println()
}

// Example 5: Context with values (for demonstration)
func exampleContextWithValues() {
	fmt.Println("5. Context with Values Example:")
	fmt.Println("   Using context to pass metadata...")

	// Create context with values
	ctx := context.WithValue(context.Background(), "userID", "12345")
	ctx = context.WithValue(ctx, "requestID", "req-abc-123")

	// Use the context
	resp, err := rss.Read(ctx, "https://httpbin.org/get", false)
	if err != nil {
		fmt.Printf("   Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// Parse the response
	channel, err := rss.Regular(ctx, resp)
	if err != nil {
		fmt.Printf("   Parse error: %v\n", err)
		return
	}

	fmt.Printf("   Successfully processed feed: %s\n", channel.Title)
	fmt.Printf("   User ID from context: %v\n", ctx.Value("userID"))
	fmt.Printf("   Request ID from context: %v\n", ctx.Value("requestID"))
	fmt.Println()
}
