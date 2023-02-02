package vct

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	listenAddr       = "127.0.0.1:8080"
	clientReqTimeout = time.Second * 10
)

var (
	l            = log.New(os.Stderr, "vcv: ", log.Ldate|log.Ltime|log.LUTC|log.Lshortfile)
	please       = []byte("")
	errTimeout   = errors.New("failed to fetch config in time")
	errFailedAtt = errors.New("failed to get attestation document")
)

// ConfigViewer implements an HTTP API that's meant to return the body of a
// third-party, remote endpoint.
type ConfigViewer struct {
	config chan []byte
	cache  *cache
	cli    *http.Client
	req    *http.Request
}

func NewConfigViewer(cli *http.Client, req *http.Request, cacheTime time.Duration) *ConfigViewer {
	return &ConfigViewer{
		config: make(chan []byte),
		cache:  newCache(cacheTime),
		cli:    cli,
		req:    req,
	}
}

func (c *ConfigViewer) Serve() {
	http.HandleFunc("/verify", c.verifyHandler)
	go c.provideConfig()
	l.Printf("Listening on %s.", listenAddr)
	l.Fatal(http.ListenAndServe(listenAddr, nil))
}

func (c *ConfigViewer) tryWhileErr(f func() ([]byte, error)) []byte {
	config, err := f()

	for err != nil {
		l.Printf("Failed to fetch config: %v", err)
		time.Sleep(time.Second * 30)
		config, err = f()
	}

	return config
}

func (c *ConfigViewer) fetchConfig() ([]byte, error) {
	var resp *http.Response

	resp, err := c.cli.Do(c.req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("got status code %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	l.Print("Successfully fetched config from backend.")

	return body, nil
}

func (c *ConfigViewer) updateCache() {
	body := c.tryWhileErr(c.fetchConfig)
	c.cache.update(body, time.Now().UTC())
	l.Printf("Updated cache with %d-byte config.", len(body))
}

func (c *ConfigViewer) provideConfig() {
	l.Print("Starting loop to provide config.")
	c.updateCache()
	for {
		select {
		case <-time.After(time.Minute * 5):
			// It's time to update our cached config.  This function blocks
			// until our cache is updated.
			c.updateCache()
		case _ = <-c.config:
			// A client requested the config.  Serve a cached copy.
			c.config <- c.cache.get()
			l.Print("Sent config to client.")
		}
	}
}

func (c *ConfigViewer) verifyHandler(w http.ResponseWriter, r *http.Request) {
	// Request a cached copy of the config.
	c.config <- please

	// Get an attestation document if the client provided a nonce.
	if nonce := r.URL.Query().Get("nonce"); nonce != "" {
		resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:8443/attestation?nonce=%s", nonce))
		if err != nil {
			http.Error(w, errFailedAtt.Error(), http.StatusInternalServerError)
			return
		}
		doc, err := io.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, errFailedAtt.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Add("X-Attestation-Document", string(doc))
	}

	// Either send the config or time out; whatever comes first.
	select {
	case config := <-c.config:
		fmt.Fprintf(w, string(config))
	case <-time.After(clientReqTimeout):
		http.Error(w, errTimeout.Error(), http.StatusInternalServerError)
	}
}
