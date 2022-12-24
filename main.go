package main

import (
	"flag"
	"fmt"
	"github.com/briandowns/openweathermap"
	"github.com/mmcdole/gofeed"
	log "github.com/sirupsen/logrus"
	"os"
)

func main() {
	verbose := flag.Bool("verbose", false, "show extended output")
	feedUrl := flag.String("feed", "https://tass.ru/rss/v2.xml", "News feed (RSS, Atom)")
	weatherApiKey := flag.String("weather-api-key", "", "openweathermap.org API key, required")
	weatherLanguage := flag.String("weather-language", "ru", "weather display language")
	weatherUnits := flag.String("weather-units", "C", "weather measurement system, one of: C, F, K")
	weatherLocation := flag.String("weather-location", "Pskov, Russia", "weather location name")
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
	if len(*feedUrl) == 0 {
		log.Fatal("feed url required")
	}
	log.Info("read feed")
	feedParser := gofeed.NewParser()
	feed, err := feedParser.ParseURL(*feedUrl)
	if err != nil {
		log.Fatalf("unable to read feed: %s", err)
	}
	for _, item := range feed.Items {
		log.Infof("feed item: published=%s, title: %s, description: %s", item.PublishedParsed.Format("15:04:05 02.01.2006"), item.Title, item.Description)
	}

	//weather
	if len(*weatherApiKey) == 0 {
		log.Fatal("openweathermap.org API key required")
	}
	log.Info("read weather")

	currentWeather, err := openweathermap.NewCurrent(*weatherUnits, *weatherLanguage, *weatherApiKey)
	if err != nil {
		log.Fatalf("unable to get current weather: %s", err)
	}
	//if err := currentWeather.CurrentByName(*weatherLocation); err != nil {
	//	log.Fatalf("unable to get current weather by location: %s", err)
	//}

	//TODO
	fmt.Println(weatherLocation)
	fmt.Println(currentWeather)

	//draw
	//TODO
}
