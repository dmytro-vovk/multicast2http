/**
 * @author Dmitry Vovk <dmitry.vovk@gmail.com>
 * @copyright 2014
 */
package main

import (
	"os/signal"
	"log"
	"os"
	"time"
	"syscall"
	"conf"
	"net/http"
	"flag"
	"fmt"
	"runtime"
	"stream"
)

var (
	urlsConfigPath = flag.String("sources", "../config/urls.json", "File with URL to source mappgings")
	listenOn       = flag.String("listen", ":7979", "Ip:port to listen for clients")
	urls conf.UrlConfig
)

/**
 * Handler to initiate streaming (or not)
 */
func urlHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Connection from %s", r.RemoteAddr)
	if url, has := urls[r.URL.Path]; has {
		log.Printf("Serving source %s", url.Source)
		doStream(w, url)
		log.Printf("Stream ended")
	} else {
		log.Printf("Source not found for URL %s", r.URL.Path)
		notFound(w)
	}
}

/**
 * 404 Not Found page
 */
func notFound(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(404)
	fmt.Fprintln(w, "<h1>Not found</h1><p>Requested source not defined</p>")
}

/**
 * 500 Internal Error page
 */
func serverFail(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(500)
	fmt.Fprintln(w, "<h1>Internal Error</h1><p>"+message+"</p>")
}

/**
 * OS signals listener
 */
func osListener() {
	osExitSignals := make(chan os.Signal, 1)
	osHupSignals := make(chan os.Signal, 1)
	signal.Notify(osExitSignals, os.Interrupt, os.Kill)
	signal.Notify(osHupSignals, syscall.SIGHUP)
	for {
		select {
		case signal := <-osExitSignals:
			log.Fatalf("Exiting due to %s", signal)
		case <-osHupSignals:
			go loadUrlConfig()
		default:
			time.Sleep(100 * time.Millisecond)
		}
		time.Sleep(50 * time.Millisecond)
	}
}

/**
 * Run streaming for given URL
 */
func doStream(w http.ResponseWriter, url conf.Url) {
	c, err := stream.GetStreamSource(url)
	if err != nil {
		serverFail(w, "Could not get stream source")
		return
	}
	b := make([]byte, 1472) // Length of UDP packet payload
	localAddress := c.LocalAddr().String()
	for {
		n, _, err := c.ReadFrom(b)
		if err != nil {
			log.Printf("Failed to read from stream: %s", err)
			return
		}
		if url.Source == localAddress {
			w.Write(b[:n])
		}
	}
}

/**
 * Reread sources config
 */
func loadUrlConfig() {
	cfg, err := conf.ReadUrls(urlsConfigPath)
	if err == nil {
		reconfigureSources(cfg)
	} else {
		log.Print("Config not loaded")
	}
}

/**
 * Apply new config, remove non-existing sources, add new ones
 */
func reconfigureSources(newConfig conf.UrlConfig) {
	urls = newConfig
}

/**
 * Main entry point
 */
func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.Parse()
	log.Printf("Process ID: %d", os.Getpid())
	loadUrlConfig()
	go osListener()
	http.HandleFunc("/", urlHandler)
	log.Printf("Listening on %s", *listenOn)
	log.Fatalf("%s", http.ListenAndServe(*listenOn, nil))
}
