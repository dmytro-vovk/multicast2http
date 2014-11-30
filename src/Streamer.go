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
	urlsConfigPath     = flag.String("sources", "config/sources.json", "File with URL to source mappings")
	networksConfigPath = flag.String("networks", "config/networks.json", "File with networks to sets mappings")
	listenOn           = flag.String("listen", ":7979", "Ip:port to listen for clients")
	hlsDir             = flag.String("hls-dir", "", "Directory to store HLS streams")
	coder              = flag.String("ffmpeg", "", "Path to ffmpeg executable")
	fakeStream         = flag.String("fake-stream", "fake.ts", "Fake stream to return to non authorized clients")
	enableWebControls  = flag.Bool("enable-web-controls", false, "Whether to enable controls via special paths")
	hlsChunkLen        = flag.Int("hls-chunk-len", 10, "Length of HLS chunk in seconds")
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.Parse()
	log.Printf("Process ID: %d", os.Getpid())
	conf.LoadConfig(*urlsConfigPath, *networksConfigPath)
	conf.FakeStream = *fakeStream
	conf.HlsChunkLen = *hlsChunkLen
}

// Main entry point
func main() {
	if *enableWebControls {
		http.HandleFunc("/server-status", response.ShowStatus)
		http.HandleFunc("/reload-config", server.ReloadConfigs)
	}
	if err := hls.SetupHLS(*hlsDir, *coder); err == nil {
		hls.RunStreams(conf.Urls)
		http.HandleFunc("/", hls.UrlHandler)
	} else {
		http.HandleFunc("/", stream.UrlHandler)
	}
	http.HandleFunc("/streams.json", server.StreamList)
	log.Printf("Listening on %s", *listenOn)
	log.Fatalf("%s", http.ListenAndServe(*listenOn, nil))
}
