package stream

import (
	"auth"
	"conf"
	"log"
	"net/http"
	netUrl "net/url"
	"response"
)

// Handler to initiate streaming (or not)
func UrlHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Connection from %s", r.RemoteAddr)
	// Disable keep-alive
	w.Header().Set("Connection", "close")
	if url, ok := conf.Urls[r.URL.Path]; ok {
		if auth.CanAccess(url, r.RemoteAddr) {
			log.Printf("Serving source %s at %s", r.URL.Path, url.Source)
			// Track number of connected users
			conf.StatsChannel <- true
			defer func() {
				conf.StatsChannel <- false
			}()
			// Source address will be parsed for sure here. @see conf.configValid()
			parsedUrl, _ := netUrl.Parse(url.Source)
			if parsedUrl.Scheme == "udp" {
				url.Source = parsedUrl.Host
				UdpStream(w, url)
			} else if parsedUrl.Scheme == "http" {
				HttpStream(w, url)
			} else {
				log.Printf("Unsupported stream protocol: ", parsedUrl.Scheme)
				response.NotFound(w)
				return
			}
			log.Printf("Stream ended")
		} else {
			log.Printf("User at %s cannot access %s", r.RemoteAddr, url.Source)
			http.ServeFile(w, r, conf.FakeStream)
		}
	} else {
		log.Printf("Source not found for URL %s", r.URL.Path)
		response.NotFound(w)
	}
}
