/**
 * @author Dmitry Vovk <dmitry.vovk@gmail.com>
 * @copyright 2014
 */
package main

import (
	"os/signal"
	"log"
	"os"
	"syscall"
	"conf"
	"net/http"
	"flag"
	"runtime"
	"stream"
	"response"
	"net"
)

var (
	urlsConfigPath     = flag.String("sources", "../config/urls.json", "File with URL to source mappgings")
	networksConfigPath = flag.String("networks", "../config/networks.json", "File with networks to sets mappings")
	listenOn           = flag.String("listen", ":7979", "Ip:port to listen for clients")
	fakeStream         = flag.String("fake-stream", "fake.ts", "Fake stream to return to non authorized clients")
	enableWebControls  = flag.Bool("enable-web-controls", false, "Wether to enable controls via special paths")
	urls conf.UrlConfig
	networks conf.NetworkConfig
statsChannel chan bool
)

/**
 * Handler to initiate streaming (or not)
 */
func urlHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Connection from %s", r.RemoteAddr)
	if url, has := urls[r.URL.Path]; has {
		if canAccess(url, r.RemoteAddr) {
			log.Printf("Serving source %s", url.Source)
			statsChannel <- true
			defer func() {
				statsChannel <- false
			}()
			doStream(w, url)
			log.Printf("Stream ended")
		} else {
			log.Printf("User at %s cannot access %s", r.RemoteAddr, url.Source)
			http.ServeFile(w, r, *fakeStream)
		}
	} else {
		log.Printf("Source not found for URL %s", r.URL.Path)
		response.NotFound(w)
	}
}

/**
 * Tells ir remote address allowed to access particular URL
 */
func canAccess(url conf.Url, address string) bool {
	host, _, _ := net.SplitHostPort(address)
	ip := net.ParseIP(host)
	for _, v := range networks {
		if v.Network.Contains(ip) {
			for _, set := range v.Sets {
				if set == url.Set {
					return true
				}
			}
		}
	}
	return false
}

/**
 * OS stop signals listener
 */
func osListener() {
	osExitSignals := make(chan os.Signal, 1)
	signal.Notify(osExitSignals, os.Interrupt, os.Kill)
	signal := <-osExitSignals
	log.Fatalf("Exiting due to %s", signal)
}

/**
 * OS HUP signal listener
 */
func hupListener() {
	osHupSignals := make(chan os.Signal, 1)
	signal.Notify(osHupSignals, syscall.SIGHUP)
	for {
		<-osHupSignals
		go loadUrlConfig()
		go loadNetworkConfig()
	}
}

/**
 * Run streaming for given URL
 */
func doStream(w http.ResponseWriter, url conf.Url) {
	c, err := stream.GetStreamSource(url)
	if err != nil {
		response.ServerFail(w, "Could not get stream source")
		return
	}
	defer c.Close()
	b := make([]byte, 1472) // Length of UDP packet payload
	localAddress := c.LocalAddr().String()
	for {
		n, _, err := c.ReadFrom(b)
		if err != nil {
			log.Printf("Failed to read from stream: %s", err)
			return
		}
		if url.Source == localAddress {
			if _, err := w.Write(b[:n]); err != nil {
				return
			}
		}
	}
}

/**
 * Reread sources config
 */
func loadUrlConfig() {
	cfg, err := conf.ReadUrls(urlsConfigPath)
	if err == nil {
		urls = cfg
	} else {
		log.Print("Config not loaded")
	}
}

func loadNetworkConfig() {
	cfg, err := conf.ReadNetworks(networksConfigPath)
	if err == nil {
		networks = cfg
	} else {
		log.Print("Network config not loaded")
	}
}

func reloadConfigs(w http.ResponseWriter, r *http.Request) {
	loadUrlConfig()
	loadNetworkConfig()
	response.ConfigReloaded(w)
}

/**
 * Stats listener
 */
func statsCollector() {
	for {
		if s := <-statsChannel; s {
			response.Stats.RunningStreams++
		} else if response.Stats.RunningStreams > 0 { // Just to prevent uint underflow
			response.Stats.RunningStreams--
		}
	}
}

// Main entry point
func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.Parse()
	log.Printf("Process ID: %d", os.Getpid())
	loadUrlConfig()
	loadNetworkConfig()
	statsChannel = make(chan bool, 10)
	go statsCollector()
	go osListener()
	go hupListener()
	if *enableWebControls {
		http.HandleFunc("/server-status", response.ShowStatus)
		http.HandleFunc("/reload-config", reloadConfigs)
	}
	http.HandleFunc("/", urlHandler)
	log.Printf("Listening on %s", *listenOn)
	log.Fatalf("%s", http.ListenAndServe(*listenOn, nil))
}
