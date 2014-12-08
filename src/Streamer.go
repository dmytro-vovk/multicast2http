/**
 * @author Dmitry Vovk <dmitry.vovk@gmail.com>
 * @copyright 2014
 */
package main

import (
	"conf"
	"flag"
	"hls"
	"log"
	"net/http"
	"os"
	"response"
	"runtime"
	"server"
	"stream"
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
		hls.RunStreams(conf.Urls)
		http.HandleFunc("/channels.json", hls.ChannelsListHandler)
		http.HandleFunc("/crossdomain.xml", hls.CrossDomainXmlHandler)
		http.HandleFunc("/", hls.UrlHandler)
	} else {
		http.HandleFunc("/", stream.UrlHandler)
	}
	log.Printf("Listening on %s", conf.Conf().Listen)
	log.Fatalf("%s", http.ListenAndServe(conf.Conf().Listen, nil))
}
