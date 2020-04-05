package main

import (
	"bufio"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"

	"github.com/ungerik/go-rss"
)

func main() {
	// Read file line by line, see https://stackoverflow.com/a/16615559/2777965
	file, err := os.Open("list.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		stringURL := scanner.Text()
		ext := filepath.Ext(stringURL)

		u, err := url.Parse(stringURL)
		if err != nil {
			panic(err)
		}

		fmt.Println("\n" + u.Host)

		reddit := false
		if u.Host == "reddit.com" {
			reddit = true
		}

		resp, err := rss.Read(stringURL, reddit)
		if err != nil {
			fmt.Println(err)
		}

		if ext == ".atom" || u.Host == "reddit.com" {
			feed, err := rss.Atom(resp)
			if err != nil {
				fmt.Println(err)
			}

			for _, entry := range feed.Entry {
				fmt.Println(entry.Updated + " " + entry.Title)
			}
		} else {
			channel, err := rss.Regular(resp)
			if err != nil {
				fmt.Println(err)
			}

			fmt.Println(channel.Title)

			for _, item := range channel.Item {
				time, err := item.PubDate.Parse()
				if err != nil {
					fmt.Println(err)
				}
				fmt.Println(time.String() + " " + item.Title + " " + item.Link)
			}
		}
	}
}
