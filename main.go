package main

import (
	"flag"
	log "github.com/sirupsen/logrus"
	"os"
)

func main() {
	verbose := flag.Bool("verbose", false, "show extended output")
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
	//TODO

	//weather
	//TODO

	//draw
	//TODO
}
