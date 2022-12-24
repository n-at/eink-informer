package main

import (
	"flag"
	"github.com/mmcdole/gofeed"
	log "github.com/sirupsen/logrus"
	"os"
)

func main() {
	verbose := flag.Bool("verbose", false, "show extended output")
	feedUrl := flag.String("feed", "https://tass.ru/rss/v2.xml", "New feed (RSS, Atom)")
	flag.Parse()

	//prepare logger
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	log.SetOutput(os.Stdout)
	if *verbose {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	//rss
	log.Info("reading feed")
	feedParser := gofeed.NewParser()
	feed, err := feedParser.ParseURL(*feedUrl)
	if err != nil {
		log.Fatalf("unable to read feed: %s", err)
	}
	for _, item := range feed.Items {
		log.Infof("feed item: published=%s, title: %s, description: %s", item.PublishedParsed.Format("15:04:05 02.01.2006"), item.Title, item.Description)
	}

	//weather
	//TODO

	//draw
	//TODO
}
