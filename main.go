package main

import (
	"flag"
	"fmt"
	"github.com/briandowns/openweathermap"
	"github.com/fogleman/gg"
	"github.com/mmcdole/gofeed"
	log "github.com/sirupsen/logrus"
	"image/color"
	"os"
	"time"
)

const (
	ImageWidth  = 800
	ImageHeight = 480

	Padding     = 5.0
	DatePadding = 10.0

	WeatherStartX = 0.0
	WeatherWidth  = ImageWidth / 2.0

	FeedStartX  = WeatherWidth + Padding
	FeedWidth   = ImageWidth - FeedStartX
	FeedPadding = 15.0
)

func main() {
	verbose := flag.Bool("verbose", false, "show extended output")
	feedUrl := flag.String("feed", "https://tass.ru/rss/v2.xml", "News feed (RSS, Atom), required")
	weatherApiKey := flag.String("weather-api-key", "", "openweathermap.org API key, required")
	weatherLanguage := flag.String("weather-language", "ru", "weather display language")
	weatherUnits := flag.String("weather-units", "C", "weather measurement system, one of: C, F, K")
	weatherLocation := flag.String("weather-location", "Pskov, Russia", "weather location name")
	imageOutputPath := flag.String("output", "output.png", "image output path, required")
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

	fontHeading, err := gg.LoadFontFace("fonts/Roboto_Condensed/RobotoCondensed-Bold.ttf", 28)
	if err != nil {
		log.Fatalf("unable to load heading font: %s", err)
	}
	fontFeedHeader, err := gg.LoadFontFace("fonts/Roboto_Condensed/RobotoCondensed-Bold.ttf", 16)
	if err != nil {
		log.Fatalf("unable to load feed header font: %s", err)
	}
	fontFeedText, err := gg.LoadFontFace("fonts/Roboto_Condensed/RobotoCondensed-Regular.ttf", 14)
	if err != nil {
		log.Fatalf("unable to load feed text font: %s", err)
	}

	//draw
	ctx := gg.NewContext(ImageWidth, ImageHeight)

	ctx.DrawRectangle(0, 0, ImageWidth, ImageHeight)
	ctx.SetColor(color.White)
	ctx.Fill()

	ctx.DrawLine(WeatherWidth, 0, WeatherWidth, ImageHeight)
	ctx.SetColor(color.Black)
	ctx.SetLineWidth(1)
	ctx.Stroke()

	//current date and time
	currentDate := time.Now().Format("02.01.2006 15:04")
	ctx.SetFontFace(fontHeading)
	ctx.SetColor(color.Black)
	w, h := ctx.MeasureString(currentDate)
	ctx.DrawString(currentDate, ImageWidth-w-10, h)

	//feed
	currentY := h + DatePadding + FeedPadding
	for itemIdx := 0; itemIdx < len(feed.Items); itemIdx++ {
		item := feed.Items[itemIdx]

		ctx.SetFontFace(fontFeedHeader)
		header := fmt.Sprintf("%s %s", item.PublishedParsed.Format("15:04:05 02.01.2006"), item.Title)
		headerWrapped := ctx.WordWrap(header, FeedWidth)
		for _, headerLine := range headerWrapped {
			_, h := ctx.MeasureString(headerLine)
			ctx.DrawString(headerLine, FeedStartX, currentY)
			currentY += h
		}

		ctx.SetFontFace(fontFeedText)
		textWrapped := ctx.WordWrap(item.Description, FeedWidth)
		for _, textLine := range textWrapped {
			_, h := ctx.MeasureString(textLine)
			ctx.DrawString(textLine, FeedStartX, currentY)
			currentY += h
		}

		currentY -= 5.0

		ctx.DrawLine(FeedStartX, currentY, ImageWidth, currentY)
		ctx.Stroke()

		currentY += FeedPadding
		if currentY >= ImageHeight {
			break
		}
	}

	if err := ctx.SavePNG(*imageOutputPath); err != nil {
		log.Fatalf("unable to save output image: %s", err)
	}
}
