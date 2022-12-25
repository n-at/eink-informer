# eink-informer

Creates image with information for e-ink display.
Prints RSS feed and weather forecast on image that can be output on e-ink display via [go-eink](https://github.com/n-at/go-eink). 

## Build

Go 1.19+ required.

```bash
go build -a -o app
```

## Usage

```txt
-feed string
    News feed (RSS, Atom), required (default "https://tass.ru/rss/v2.xml")
-feed-content-max-length int
    maximum length of feed content text (default 150)
-feed-title-max-length int
    maximum length of feed item title (default 100)
-output string
    image output path, required (default "output.png")
-verbose
    show extended output
-weather-api-key string
    openweathermap.org API key, required
-weather-language string
    weather display language (default "ru")
-weather-location string
    weather location name (default "Pskov, Russia")
-weather-units string
    weather measurement system, one of: C, F, K (default "C")
-weather-timezone-offset float
    timezone offset, hours (default 3)
```

## Uses

* [sirupsen/logrus](https://github.com/sirupsen/logrus)
* [fogleman/gg](https://github.com/fogleman/gg)
* [briandowns/openweathermap](https://github.com/briandowns/openweathermap)
* [mmcdole/gofeed](https://github.com/mmcdole/gofeed)
* Icons by [icons8.com](https://icons8.com)
