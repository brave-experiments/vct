package main

import (
	"log"
	"net/http"
	"time"

	"github.com/brave-experiments/vcv"
)

func main() {
	req, err := http.NewRequest("GET", "https://nymity.ch", nil)
	if err != nil {
		log.Fatal(err)
	}

	c := vcv.NewConfigViewer(&http.Client{}, req, time.Minute*10)
	c.Serve()
}
