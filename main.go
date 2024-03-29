package main

import (
	"flag"
	"fmt"
	owm "github.com/briandowns/openweathermap"
	"github.com/fogleman/gg"
	"github.com/mmcdole/gofeed"
	log "github.com/sirupsen/logrus"
	"golang.org/x/image/font"
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

	WeatherStartX         = 0.0
	WeatherWidth          = ImageWidth / 2.0
	WeatherBigIconSize    = 60
	WeatherSmallIconSize  = 30
	WeatherPadding        = 10
	WeatherForecastWidth  = 50
	WeatherForecastHeight = 100

	FeedStartX  = WeatherWidth + Padding
	FeedWidth   = ImageWidth - FeedStartX
	FeedPadding = 15.0
)

var (
	fontHeading           font.Face
	fontFeedHeader        font.Face
	fontFeedText          font.Face
	fontWeatherCurrent    font.Face
	fontWeatherConditions font.Face
	fontWeatherForecast   font.Face

	iconUnknownBig    image.Image
	iconUnknownSmall  image.Image
	weatherIconsBig   = make(map[string]image.Image)
	weatherIconsSmall = make(map[string]image.Image)

	timezoneOffset   = +3.0
	temperatureUnits = "C"
)

type weather struct {
	iconBig    image.Image
	iconSmall  image.Image
	conditions string
	tempCur    string
	tempRange  string
	humidity   int
	date       string
	time       string
}

func main() {
	verbose := flag.Bool("verbose", false, "show extended output")
	feedUrl := flag.String("feed", "https://tass.ru/rss/v2.xml", "News feed (RSS, Atom), required")
	feedTitleMaxLength := flag.Int("feed-title-max-length", 100, "maximum length of feed item title")
	feedContentMaxLength := flag.Int("feed-content-max-length", 150, "maximum length of feed content text")
	weatherApiKey := flag.String("weather-api-key", "", "openweathermap.org API key, required")
	weatherLanguage := flag.String("weather-language", "ru", "weather display language")
	weatherUnits := flag.String("weather-units", "C", "weather measurement system, one of: C, F, K")
	weatherLocation := flag.String("weather-location", "Pskov, Russia", "weather location name")
	imageOutputPath := flag.String("output", "output.png", "image output path, required")
	offset := flag.Float64("weather-timezone-offset", 3.0, "timezone offset, hours")
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

	loadResources()
	timezoneOffset = *offset
	temperatureUnits = *weatherUnits

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
	//current
	ctx.DrawImage(weatherValueCurrent.iconBig, WeatherStartX, 0)
	//current temp/humidity
	ctx.SetFontFace(fontWeatherCurrent)
	ctx.DrawString(fmt.Sprintf("%s / %d%%", weatherValueCurrent.tempCur, weatherValueCurrent.humidity), WeatherStartX+WeatherBigIconSize+WeatherPadding, 30)
	//current conditions
	conditions := ctx.WordWrap(weatherValueCurrent.conditions, ImageWidth-WeatherStartX-WeatherBigIconSize+WeatherPadding)
	ctx.SetFontFace(fontWeatherConditions)
	ctx.DrawString(conditions[0], WeatherStartX+WeatherBigIconSize+WeatherPadding, 50)
	//forecast
	forecastCurrentX := int(WeatherStartX)
	forecastCurrentY := WeatherBigIconSize + WeatherPadding
	for _, item := range weatherValueForecast {
		ctx.DrawImage(item.iconSmall, forecastCurrentX+(WeatherForecastWidth-WeatherSmallIconSize)/2, forecastCurrentY)

		ctx.SetFontFace(fontWeatherForecast)
		dateW, dateH := ctx.MeasureString(item.date)
		x := float64(forecastCurrentX) + (WeatherForecastWidth-dateW)/2.0
		y := float64(forecastCurrentY) + WeatherSmallIconSize + WeatherPadding
		ctx.DrawString(item.date, x, y)

		ctx.SetFontFace(fontWeatherForecast)
		timeW, timeH := ctx.MeasureString(item.time)
		x = float64(forecastCurrentX) + (WeatherForecastWidth-timeW)/2.0
		y = float64(forecastCurrentY) + WeatherSmallIconSize + WeatherPadding + dateH + 5
		ctx.DrawString(item.time, x, y)

		ctx.SetFontFace(fontWeatherForecast)
		tempW, _ := ctx.MeasureString(item.tempRange)
		x = float64(forecastCurrentX) + (WeatherForecastWidth-tempW)/2.0
		y = float64(forecastCurrentY) + WeatherSmallIconSize + WeatherPadding + dateH + timeH + 10
		ctx.DrawString(item.tempRange, x, y)

		//next forecast block position
		forecastCurrentX += WeatherForecastWidth
		if forecastCurrentX+WeatherForecastWidth > WeatherWidth+WeatherStartX {
			forecastCurrentX = WeatherStartX
			forecastCurrentY += WeatherForecastHeight
		}
		if forecastCurrentY+WeatherForecastHeight > ImageHeight {
			break
		}
	}

	if err := ctx.SavePNG(*imageOutputPath); err != nil {
		log.Fatalf("unable to save output image: %s", err)
	}
}

///////////////////////////////////////////////////////////////////////////////

func loadResources() {
	var err error

	//load icons
	iconUnknownBig, err = gg.LoadImage("icons/60/unknown.png")
	if err != nil {
		log.Fatalf("unable to load big unknown image: %s", err)
	}
	iconUnknownSmall, err = gg.LoadImage("icons/30/unknown.png")
	if err != nil {
		log.Fatalf("unable to load small unknown image: %s", err)
	}

	icons := []string{
		"01d", "01n", "02d", "02n", "03d", "03n", "04d", "04n",
		"09d", "09n", "10d", "10n", "11d", "11n", "13d", "13n", "50d", "50n",
	}
	for _, icon := range icons {
		weatherIconsBig[icon], err = gg.LoadImage(fmt.Sprintf("icons/60/%s.png", icon))
		if err != nil {
			log.Fatalf("unable to load big icon %s: %s", icon, err)
		}
		weatherIconsSmall[icon], err = gg.LoadImage(fmt.Sprintf("icons/30/%s.png", icon))
		if err != nil {
			log.Fatalf("unable to load small icon %s: %s", icon, err)
		}
	}

	//load fonts
	fontHeading, err = gg.LoadFontFace("fonts/Roboto_Condensed/RobotoCondensed-Bold.ttf", 28)
	if err != nil {
		log.Fatalf("unable to load heading font: %s", err)
	}
	fontFeedHeader, err = gg.LoadFontFace("fonts/Roboto_Condensed/RobotoCondensed-Bold.ttf", 16)
	if err != nil {
		log.Fatalf("unable to load feed header font: %s", err)
	}
	fontFeedText, err = gg.LoadFontFace("fonts/Roboto_Condensed/RobotoCondensed-Regular.ttf", 14)
	if err != nil {
		log.Fatalf("unable to load feed text font: %s", err)
	}
	fontWeatherCurrent, err = gg.LoadFontFace("fonts/Roboto_Condensed/RobotoCondensed-Bold.ttf", 28)
	if err != nil {
		log.Fatalf("unable to load weather current font: %s", err)
	}
	fontWeatherConditions, err = gg.LoadFontFace("fonts/Roboto_Condensed/RobotoCondensed-Regular.ttf", 18)
	if err != nil {
		log.Fatalf("unable to load weather conditions font: %s", err)
	}
	fontWeatherForecast, err = gg.LoadFontFace("fonts/Roboto_Condensed/RobotoCondensed-Regular.ttf", 14)
	if err != nil {
		log.Fatalf("unable to load weather forecast font: %s", err)
	}
}

func extractWeatherFromCurrent(current *owm.CurrentWeatherData) weather {
	w := extractWeather(current.Weather)
	w.tempCur = formatTemperature(current.Main.Temp)
	w.tempRange = fmt.Sprintf(formatTemperatureRange(current.Main.Temp, current.Main.FeelsLike))
	w.humidity = current.Main.Humidity

	t := time.Unix(int64(current.Dt), 0).Add(time.Duration(timezoneOffset * float64(time.Hour)))
	w.time = t.Format("15:04")
	w.date = t.Format("02.01")
	return w
}

func extractWeatherFromForecast(forecast owm.Forecast5WeatherList) weather {
	w := extractWeather(forecast.Weather)
	w.tempCur = formatTemperature(forecast.Main.Temp)
	w.tempRange = fmt.Sprintf(formatTemperatureRange(forecast.Main.Temp, forecast.Main.FeelsLike))
	w.humidity = forecast.Main.Humidity

	t := forecast.DtTxt.Add(time.Duration(timezoneOffset * float64(time.Hour)))
	w.time = t.Format("15:04")
	w.date = t.Format("02.01")
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

	iconBig, ok := weatherIconsBig[icon]
	if ok {
		w.iconBig = iconBig
	} else {
		w.iconBig = iconUnknownBig
	}

	iconSmall, ok := weatherIconsSmall[icon]
	if ok {
		w.iconSmall = iconSmall
	} else {
		w.iconSmall = iconUnknownSmall
	}

	return w
}

func trimFeedText(s string, l int) string {
	if len([]rune(s)) < l {
		return s
	} else {
		return string([]rune(s)[:l]) + "..."
	}
}

func formatTemperature(value float64) string {
	round := int(math.Round(value))
	str := fmt.Sprintf("%+d", round)
	if round == 0 {
		str = "0"
	}
	return fmt.Sprintf("%s °%s", str, temperatureUnits)
}

func formatTemperatureRange(value1, value2 float64) string {
	round1 := int(math.Round(value1))
	round2 := int(math.Round(value2))

	str1 := fmt.Sprintf("%+d", round1)
	str2 := fmt.Sprintf("%+d", round2)

	if round1 == 0 {
		str1 = "0"
	}
	if round2 == 0 {
		str2 = "0"
	}

	return fmt.Sprintf("%s / %s", str1, str2)
}
