package main

import (
	"flag"
	"fmt"
	owm "github.com/briandowns/openweathermap"
	"github.com/fogleman/gg"
	"github.com/mmcdole/gofeed"
	log "github.com/sirupsen/logrus"
	"image"
	"image/color"
	"math"
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

	TimezoneOffset = +3.0
)

type weather struct {
	icon       image.Image
	conditions string
	tempMin    string
	tempMax    string
	tempCur    string
	tempRange  string
	humidity   int
	time       string
}

func main() {
	verbose := flag.Bool("verbose", false, "show extended output")
	feedUrl := flag.String("feed", "https://tass.ru/rss/v2.xml", "News feed (RSS, Atom), required")
	feedTitleMaxLength := flag.Int("feed-title-max-length", 100, "maximum length of feed item title")
	feedContentMaxLength := flag.Int("feed-content-max-length", 200, "maximum length of feed content text")
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

	//weather
	if len(*weatherApiKey) == 0 {
		log.Fatal("openweathermap.org API key required")
	}
	log.Info("read weather")
	currentWeather, err := owm.NewCurrent(*weatherUnits, *weatherLanguage, *weatherApiKey)
	if err != nil {
		log.Fatalf("unable to get current weather: %s", err)
	}
	if err := currentWeather.CurrentByName(*weatherLocation); err != nil {
		log.Fatalf("unable to get current weather by location: %s", err)
	}
	weatherForecast, err := owm.NewForecast("5", *weatherUnits, *weatherLanguage, *weatherApiKey)
	if err != nil {
		log.Fatalf("unable to get weather forecast: %s", err)
	}
	if err := weatherForecast.DailyByName(*weatherLocation, 5*8); err != nil {
		log.Fatalf("unable to get weather forecast by location: %s", err)
	}

	weatherValueCurrent := extractWeatherFromCurrent(currentWeather)
	var weatherValueForecast []weather
	for _, item := range weatherForecast.ForecastWeatherJson.(*owm.Forecast5WeatherData).List {
		weatherValueForecast = append(weatherValueForecast, extractWeatherFromForecast(item))
	}
	//TODO
	fmt.Println(weatherValueCurrent)
	fmt.Println(weatherValueForecast)

	//load fonts
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

	//draw current date and time
	currentDate := time.Now().Format("02.01.2006 15:04")
	ctx.SetFontFace(fontHeading)
	ctx.SetColor(color.Black)
	w, h := ctx.MeasureString(currentDate)
	ctx.DrawString(currentDate, ImageWidth-w-10, h)

	//draw feed
	currentY := h + DatePadding + FeedPadding
	for itemIdx := 0; itemIdx < len(feed.Items); itemIdx++ {
		item := feed.Items[itemIdx]

		ctx.SetFontFace(fontFeedHeader)
		header := fmt.Sprintf("%s %s", item.PublishedParsed.Format("15:04 02.01.2006"), trimFeedText(item.Title, *feedTitleMaxLength))
		headerWrapped := ctx.WordWrap(header, FeedWidth)
		for _, headerLine := range headerWrapped {
			_, h := ctx.MeasureString(headerLine)
			ctx.DrawString(headerLine, FeedStartX, currentY)
			currentY += h
		}

		ctx.SetFontFace(fontFeedText)
		contentWrapped := ctx.WordWrap(trimFeedText(item.Description, *feedContentMaxLength), FeedWidth)
		for _, contentLine := range contentWrapped {
			_, h := ctx.MeasureString(contentLine)
			ctx.DrawString(contentLine, FeedStartX, currentY)
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

	//draw weather
	//TODO

	if err := ctx.SavePNG(*imageOutputPath); err != nil {
		log.Fatalf("unable to save output image: %s", err)
	}
}

///////////////////////////////////////////////////////////////////////////////

func trimFeedText(s string, l int) string {
	if len([]rune(s)) < l {
		return s
	} else {
		return string([]rune(s)[:l]) + "..."
	}
}

func extractWeatherFromCurrent(current *owm.CurrentWeatherData) weather {
	w := extractWeather(current.Weather)
	w.tempMin = formatTemperature(current.Main.TempMin)
	w.tempCur = formatTemperature(current.Main.Temp)
	w.tempMax = formatTemperature(current.Main.TempMax)
	w.tempRange = fmt.Sprintf("%s...%s", w.tempMin, w.tempMax)
	w.humidity = current.Main.Humidity
	w.time = time.Unix(int64(current.Dt), 0).Add(TimezoneOffset * time.Hour).Format("15:04 02.01")
	return w
}

func extractWeatherFromForecast(forecast owm.Forecast5WeatherList) weather {
	w := extractWeather(forecast.Weather)
	w.tempMin = formatTemperature(forecast.Main.TempMin)
	w.tempCur = formatTemperature(forecast.Main.Temp)
	w.tempMax = formatTemperature(forecast.Main.TempMax)
	w.tempRange = fmt.Sprintf("%s...%s", w.tempMin, w.tempMax)
	w.humidity = forecast.Main.Humidity
	w.time = forecast.DtTxt.Add(TimezoneOffset * time.Hour).Format("15:04 02.01")
	return w
}

func extractWeather(items []owm.Weather) weather {
	w := weather{}
	icon := ""

	for _, item := range items {
		if len(item.Icon) != 0 && len(icon) == 0 {
			icon = item.Icon
		}
		if len(item.Description) == 0 {
			continue
		}
		if len(w.conditions) != 0 {
			w.conditions += ", "
		}
		w.conditions += item.Description
	}

	if len(icon) != 0 {
		img, err := gg.LoadImage(fmt.Sprintf("icons/%s.png", icon)) //TODO check icon file name
		if err != nil {
			log.Errorf("unable to load icon %s: %s", icon, err)
			img = nil
		}
		w.icon = img
	}

	return w
}

func formatTemperature(value float64) string {
	return fmt.Sprintf("%+d", int(math.Round(value)))
}
