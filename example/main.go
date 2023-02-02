package main

import (
	"log"
	"net/http"
	"time"

	"github.com/brave-experiments/vct"
)

func main() {
	req, err := http.NewRequest("GET", "https://nymity.ch/hello-world.html", nil)
	if err != nil {
		log.Fatal(err)
	}

	c := vct.NewConfigViewer(&http.Client{}, req, time.Minute*10)
	c.Serve()
}
