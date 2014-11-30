package hls

import (
	"auth"
	"conf"
	"log"
	"net"
	"net/http"
	"path/filepath"
	"response"
	"strings"
	"time"
)

func UrlHandler(w http.ResponseWriter, r *http.Request) {
	// Redirect to m3u8 file
	if !strings.HasSuffix(r.URL.Path, ".m3u8") && !strings.HasSuffix(r.URL.Path, ".ts") {
		http.Redirect(w, r, r.URL.Path+"/stream.m3u8", http.StatusFound)
		return
	}
	// Track number of connected users
	conf.StatsChannel <- true
	defer func() {
		time.Sleep(time.Duration(conf.HlsChunkLen) * time.Second)
		conf.StatsChannel <- false
	}()
	prefix := filepath.Dir(r.URL.Path)
	if url, ok := conf.Urls[prefix]; ok {
		if auth.CanAccess(url, r.RemoteAddr) {
			host, _, _ := net.SplitHostPort(r.RemoteAddr)
			log.Printf("Client %s requests file %s", host, r.RequestURI)
			http.ServeFile(w, r, HLSDir+r.RequestURI)
		} else {
			http.ServeFile(w, r, conf.FakeStream)
		}
	} else {
		log.Printf("Source not found for URL prefix %s", prefix)
		response.NotFound(w)
	}
}
