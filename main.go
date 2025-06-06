package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"golang.org/x/sync/errgroup"
)

var URLs = []struct {
	url   string
	title string
}{
	{
		url:   "https://desktop.docker.com/mac/main/arm64/appcast.xml",
		title: "Mac ARM",
	},
	{
		url:   "https://desktop.docker.com/mac/main/arm64/appcast.xml",
		title: "Mac AMD",
	},
	{
		url:   "https://desktop.docker.com/mac/main/arm64/appcast.xml",
		title: "Windows",
	},
}

type item struct {
	Title      string  `xml:"title"`
	Visibility float64 `xml:"visibility"`
}

type channel struct {
	Item []item `xml:"item"`
}

type response struct {
	Channel channel `xml:"channel"`
}

type gather struct {
	title     string
	responses []item
}

func progress(visibility float64) string {
	width := 40.0
	s := ""

	for i := 0; i < int(visibility/100.0*width); i += 1 {
		s += "■"
	}

	for i := int(visibility / 100.0 * width); i < int(width); i += 1 {
		s += " "
	}

	return s
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	var responses []gather

	eg, _ := errgroup.WithContext(context.Background())
	for _, input := range URLs {
		input := input

		eg.Go(func() error {
			res, err := http.Get(input.url)
			if err != nil {
				return err
			}

			var response response
			body, _ := io.ReadAll(res.Body)

			if err := xml.Unmarshal(body, &response); err != nil {
				return err
			}

			responses = append(responses, gather{
				title:     input.title,
				responses: response.Channel.Item,
			})

			defer res.Body.Close()
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return err
	}

	for _, response := range responses {
		fmt.Println(response.title)
		for _, item := range response.responses {
			title := strings.Split(item.Title, " ")[0]
			fmt.Printf("\t%s\t%s %.0f%%\n", title, progress(item.Visibility), item.Visibility)
		}
		fmt.Println()
	}

	return nil
}
