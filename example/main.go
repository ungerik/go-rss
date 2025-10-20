package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/ungerik/go-rss"
)

func main() {
	// Create a context with timeout for the entire operation
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Read file line by line, see https://stackoverflow.com/a/16615559/2777965
	file, err := os.Open("list.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		// Check if context is cancelled before processing each URL
		select {
		case <-ctx.Done():
			fmt.Printf("Operation cancelled: %v\n", ctx.Err())
			return
		default:
		}

		stringURL := scanner.Text()
		ext := filepath.Ext(stringURL)

		u, err := url.Parse(stringURL)
		if err != nil {
			fmt.Printf("Error parsing URL %s: %v\n", stringURL, err)
			continue
		}

		fmt.Println("\n" + u.Host)

		reddit := false
		if u.Host == "reddit.com" {
			reddit = true
		}

		// Use context-aware function with timeout
		resp, err := rss.Read(ctx, stringURL, reddit)
		if err != nil {
			fmt.Printf("Error fetching %s: %v\n", stringURL, err)
			continue
		}

		if ext == ".atom" || u.Host == "reddit.com" {
			feed, err := rss.Atom(ctx, resp)
			if err != nil {
				fmt.Printf("Error parsing Atom feed: %v\n", err)
				continue
			}

			for _, entry := range feed.Entry {
				fmt.Println(entry.Updated + " " + entry.Title)
			}
		} else {
			channel, err := rss.Regular(ctx, resp)
			if err != nil {
				fmt.Printf("Error parsing RSS feed: %v\n", err)
				continue
			}

			fmt.Println(channel.Title)

			for _, item := range channel.Item {
				time, err := item.PubDate.Parse()
				if err != nil {
					fmt.Printf("Error parsing date: %v\n", err)
				}
				fmt.Println(time.String() + " " + item.Title + " " + item.Link)
			}
		}
	}
}
