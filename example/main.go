package main

import (
	"bufio"
	"fmt"
	"log"
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
		url := scanner.Text()
		ext := filepath.Ext(url)

		resp, err := rss.Read(url)
		if err != nil {
			fmt.Println(err)
		}

		if ext == ".atom" {
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
				fmt.Println(time.String() + " " + item.Title)
			}
		}
	}
}
