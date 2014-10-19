/**
 * @author Dmitry Vovk <dmitry.vovk@gmail.com>
 * @copyright 2014
 */
package main

import (
	"conf"
	"errors"
	"flag"
	hls "hlstream"
	"log"
	"net"
	"net/http"
	netUrl "net/url"
	"os"
	"os/signal"
	"response"
	"runtime"
	"stream"
	"syscall"
)

var (
	urlsConfigPath     = flag.String("sources", "config/sources.json", "File with URL to source mappings")
	networksConfigPath = flag.String("networks", "config/networks.json", "File with networks to sets mappings")
	listenOn           = flag.String("listen", ":7979", "Ip:port to listen for clients")
	hlsDir             = flag.String("hls-dir", "", "Directory to store HLS streams")
	ffmpeg             = flag.String("ffmpeg", "", "Path to ffmpeg executable")
	fakeStream         = flag.String("fake-stream", "fake.ts", "Fake stream to return to non authorized clients")
	enableWebControls  = flag.Bool("enable-web-controls", false, "Whether to enable controls via special paths")
	urls               conf.UrlConfig
	statsChannel       chan bool
)

// Handler to initiate streaming (or not)
func udpUrlHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Connection from %s", r.RemoteAddr)
	// Disable keep-alive
	w.Header().Set("Connection", "close")
	if url, ok := urls[r.URL.Path]; ok {
		if canAccess(url, r.RemoteAddr) {
			log.Printf("Serving source %s at %s", r.URL.Path, url.Source)
			// Track number of connected users
			statsChannel <- true
			defer func() {
				statsChannel <- false
			}()
			// Source address will be parsed for sure here. @see conf.configValid()
			parsedUrl, _ := netUrl.Parse(url.Source)
			if parsedUrl.Scheme == "udp" {
				url.Source = parsedUrl.Host
				stream.UdpStream(w, url)
			} else if parsedUrl.Scheme == "http" {
				stream.HttpStream(w, url)
			} else {
				log.Printf("Unsupported stream protocol: ", parsedUrl.Scheme)
				response.NotFound(w)
				return
			}
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

// Tells ir remote address allowed to access particular URL
func canAccess(url conf.Url, address string) bool {
	host, _, _ := net.SplitHostPort(address)
	ip := net.ParseIP(host)
	for _, n := range url.Networks {
		if n.Contains(ip) {
			return true
		}
	}
	return false
}

// OS stop signals listener
func osListener() {
	osExitSignals := make(chan os.Signal, 1)
	signal.Notify(osExitSignals, os.Interrupt, os.Kill)
	signal := <-osExitSignals
	log.Fatalf("Exiting due to %s", signal)
}

// OS HUP signal listener
func hupListener() {
	osHupSignals := make(chan os.Signal, 1)
	signal.Notify(osHupSignals, syscall.SIGHUP)
	for {
		<-osHupSignals
		loadConfig()
	}
}

// Reread sources config
func loadConfig() {
	_urls, err := conf.ReadUrls(urlsConfigPath)
	if err == nil {
		_nets, err := conf.ReadNetworks(networksConfigPath)
		if err == nil {
			urls = mergeConfigs(_urls, _nets)
		} else {
			log.Print("Network config not loaded")
		}
	} else {
		log.Print("Config not loaded")
	}
}

// Populate sources with allowed networks based on sets
func mergeConfigs(_urls conf.UrlConfig, _nets conf.NetworkConfig) conf.UrlConfig {
	// Go over sources
	for u, _url := range _urls {
		// Go over networks
		for _, _net := range _nets {
			// Go over sets
			for _, set := range _net.Sets {
				if _url.Set == set {
					_url.Networks = append(_url.Networks, _net.Network)
				}
			}
		}
		_urls[u] = _url
	}
	return _urls
}

func reloadConfigs(w http.ResponseWriter, _ *http.Request) {
	loadConfig()
	response.ConfigReloaded(w)
}

// Stats listener
func statsCollector() {
	for {
		if s := <-statsChannel; s {
			response.Stats.RunningStreams++
		} else if response.Stats.RunningStreams > 0 { // Just to prevent uint underflow
			response.Stats.RunningStreams--
		}
	}
}

// Check if HLS can be used
func setupHLS() error {
	// Supplied dir should not be empty
	if *hlsDir == "" {
		return errors.New("No HLS directory specified")
	}
	// Dir must exist
	d, err := os.Stat(*hlsDir)
	if os.IsNotExist(err) {
		// Try to create it
		if err = os.MkdirAll(*hlsDir, 0777); err != nil {
			return errors.New("Directory for HLS does not exist and unable to create")
		}
		d, err = os.Stat(*hlsDir)
	} else if err != nil {
		return err
	}
	if !d.IsDir() {
		return errors.New("HLS path is file, not directory")
	}
	_, err = os.Stat(*ffmpeg)
	if os.IsNotExist(err) {
		return errors.New("ffmpeg not found")
	}
	return nil
}

func hlsUrlHandler(w http.ResponseWriter, r *http.Request) {
	// Track number of connected users
	statsChannel <- true
	defer func() {
		statsChannel <- false
	}()
	log.Printf("Serving file %s", *hlsDir+r.RequestURI)
	http.ServeFile(w, r, *hlsDir+r.RequestURI)
}

// Main entry point
func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.Parse()
	log.Printf("Process ID: %d", os.Getpid())
	loadConfig()
	statsChannel = make(chan bool, 10)
	go statsCollector()
	go osListener()
	go hupListener()
	if *enableWebControls {
		http.HandleFunc("/server-status", response.ShowStatus)
		http.HandleFunc("/reload-config", reloadConfigs)
	}
	if err := setupHLS(); err == nil {
		hls.HLSDir = *hlsDir
		hls.Ffmpeg = *ffmpeg
		hls.RunStreams(urls)
		http.HandleFunc("/", hlsUrlHandler)
	} else {
		http.HandleFunc("/", udpUrlHandler)
	}
	log.Printf("Listening on %s", *listenOn)
	log.Fatalf("%s", http.ListenAndServe(*listenOn, nil))
}
