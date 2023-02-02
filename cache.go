package vct

import (
	"sync"
	"time"
)

type cache struct {
	sync.RWMutex
	expiry    time.Duration
	lastFetch time.Time
	lastBody  []byte
}

func newCache(expiry time.Duration) *cache {
	return &cache{
		expiry: expiry,
	}
}

func (c *cache) isValid() bool {
	c.RLock()
	defer c.RUnlock()

	if c.lastFetch.IsZero() {
		return false
	}

	now := time.Now().UTC()
	return !now.Add(-c.expiry).After(c.lastFetch)
}

func (c *cache) update(body []byte, fetch time.Time) {
	c.Lock()
	defer c.Unlock()

	c.lastBody = body
	c.lastFetch = fetch
}

func (c *cache) get() []byte {
	c.RLock()
	defer c.RUnlock()

	return c.lastBody
}
