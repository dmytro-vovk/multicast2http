package hls

import (
	"cache"
	"conf"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"response"
	"strings"
)

func UrlHandler(w http.ResponseWriter, r *http.Request) {
	// Redirect to m3u8 file
	if !strings.HasSuffix(r.URL.Path, ".m3u8") && !strings.HasSuffix(r.URL.Path, ".ts") {
		http.Redirect(w, r, r.URL.Path+"/stream.m3u8", http.StatusFound)
		return
	}
	prefix := filepath.Dir(r.URL.Path)
	if _, ok := conf.Conf().Urls[prefix]; ok {
		deliverFile(w, HLSDir+r.RequestURI)
	} else {
		log.Printf("Source not found for URL prefix %s", prefix)
		response.NotFound(w)
	}
}

func deliverFile(w http.ResponseWriter, fileName string) {
	if strings.HasSuffix(fileName, ".ts") {
		deliverTs(w, fileName)
	} else if strings.HasSuffix(fileName, ".m3u8") {
		deliverPlaylist(w, fileName)
	} else {
		w.Header().Set("Content-Type", "application/octet-stream")
		b, err := ioutil.ReadFile(fileName)
		if err != nil {
			response.NotFound(w)
		}
		w.Write(b)
	}
}

// we do not cache playlists
func deliverPlaylist(w http.ResponseWriter, fileName string) {
	b, err := ioutil.ReadFile(fileName)
	if err != nil {
		response.NotFound(w)
		return
	}
	w.Header().Set("Content-Type", "application/x-mpegurl")
	w.Write(b)
}

// cache TS files
func deliverTs(w http.ResponseWriter, fileName string) {
	b, err := cache.Cache.Get(fileName)
	if err != nil {
		response.NotFound(w)
		return
	}
	w.Header().Set("Content-Type", "video/mp2t")
	w.Write(b)
}

// output json list of available streams
func ChannelsListHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/json; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(streamsList)
}

func CrossDomainXmlHandler(w http.ResponseWriter, _ *http.Request) {
	xDomain := `<?xml version="1.0"?>
<!DOCTYPE cross-domain-policy SYSTEM "http://www.macromedia.com/xml/dtds/cross-domain-policy.dtd">
<cross-domain-policy>
<allow-access-from domain="` + conf.Conf().Web.AllowOrigin + `" />
</cross-domain-policy>`
	w.Header().Set("Content-Type", "text/xml")
	fmt.Fprintf(w, xDomain)
}
