package hls

import (
	"auth"
	"conf"
	"encoding/json"
	"fmt"
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

// output json list of available streams
func ChannelsListHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/json; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", conf.AllowDomain)
	json.NewEncoder(w).Encode(streamsList)
}

func CrossDomainXmlHandler(w http.ResponseWriter, _ *http.Request) {
	xDomain := `<?xml version="1.0"?>
<!DOCTYPE cross-domain-policy SYSTEM "http://www.macromedia.com/xml/dtds/cross-domain-policy.dtd">
<cross-domain-policy>
<allow-access-from domain="` + conf.AllowDomain + `" />
</cross-domain-policy>`
	w.Header().Set("Content-Type", "text/xml")
	fmt.Fprintf(w, xDomain)
}
