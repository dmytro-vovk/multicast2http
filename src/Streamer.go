/**
 * @author Dmitry Vovk <dmitry.vovk@gmail.com>
 * @copyright 2014
 */
package main

import (
	"assets"
	"cache"
	"conf"
	"flag"
	"hls"
	"log"
	"net"
	"net/http"
	"os"
	"response"
	"runtime"
	"server"
	"stream"
	"time"
)

var (
	configFile = flag.String("config", "config/config.json", "Path to config file")
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.Parse()
	log.Printf("Process ID: %d", os.Getpid())
	conf.LoadConfig(*configFile)
}

// Main entry point
func main() {
	if conf.Conf().Web.EnableControls {
		http.HandleFunc("/server-status", response.ShowStatus)
		http.HandleFunc("/reload-config", server.ReloadConfigs)
	}
	if err := hls.SetupHLS(conf.Conf().Hls.Dir, conf.Conf().Hls.Ffmpeg); err == nil {
		if conf.Conf().Cache.Every != 0 {
			cache.Every = time.Second * time.Duration(conf.Conf().Cache.Every)
		}
		if conf.Conf().Cache.Expires != 0 {
			cache.Expires = time.Second * time.Duration(conf.Conf().Cache.Expires)
		}
		hls.RunStreams(conf.Conf().Urls)
		http.HandleFunc("/channels.json", hls.ChannelsListHandler)
		http.HandleFunc("/crossdomain.xml", hls.CrossDomainXmlHandler)
		http.HandleFunc("/player.js", assets.HandlerJs)
		http.HandleFunc("/player.swf", assets.HandlerSwf)
		http.HandleFunc("/", logger(hls.UrlHandler))
	} else {
		http.HandleFunc("/", stream.UrlHandler)
	}
	log.Printf("Listening on %s", conf.Conf().Listen)
	log.Fatalf("%s", http.ListenAndServe(conf.Conf().Listen, nil))
}

func logger(handler func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.RequestURI == "/" {
			if ua := r.UserAgent(); ua != "" {
				log.Printf("User Agent: %s", ua)
			}
			assets.HandlerHtml(w, r)
			return
		}
		start := time.Now()
		handler(w, r)
		host, _, _ := net.SplitHostPort(r.RemoteAddr)
		log.Printf("%s %s %s %s", host, r.Method, r.RequestURI, time.Since(start))
	})
}
