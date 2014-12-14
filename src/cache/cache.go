package cache

import (
	"io/ioutil"
	"log"
	"time"
)

type cacheEntry struct {
	content    []byte
	lastAccess time.Time
}

type Cacher struct {
	files map[string]cacheEntry
}

var (
	Cache   *Cacher
	Expires = time.Second * 120
	Every   = time.Second * 5
)

func init() {
	Cache = &Cacher{
		files: make(map[string]cacheEntry, 100),
	}
	go Cache.cleaner()
}

func (c *Cacher) Get(fileName string) ([]byte, error) {
	if file, ok := c.files[fileName]; ok {
		log.Printf("Cache hit: %s", fileName)
		return file.content, nil
	}
	log.Printf("Cache miss: %s", fileName)
	b, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	c.files[fileName] = cacheEntry{
		content:    b,
		lastAccess: time.Now(),
	}
	return b, nil
}

func (c *Cacher) cleaner() {
	for {
		time.Sleep(Every)
		for fileName, file := range c.files {
			if time.Since(file.lastAccess) > Expires {
				log.Printf("Cache purge: %s", fileName)
				delete(c.files, fileName)
			}
		}
	}
}
