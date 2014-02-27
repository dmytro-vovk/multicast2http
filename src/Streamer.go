package main

import (
	"os/signal"
	"log"
	"os"
	"time"
	"errors"
	"syscall"
	"conf"
	"net/http"
	"flag"
	"fmt"
	"runtime"
	"net"
	"strconv"
	"code.google.com/p/go.net/ipv4"
)

var (
	urlsConfigPath = flag.String("sources", "../config/urls.json", "File with URL to source mappgings")
	listenOn       = flag.String("listen", ":7979", "Ip:port to listen for clients")
	urls conf.UrlConfig
)

func urlHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Connection from %s", r.RemoteAddr)
	if url, has := urls[r.URL.Path]; has {
		log.Printf("Serving source %s", url.Source)
		stream(w, url)
		log.Printf("Stream ended")
	} else {
		log.Printf("Source not found for URL %s", r.URL.Path)
		notFound(w)
	}
}

func notFound(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(404)
	fmt.Fprintln(w, "<h1>Not found</h1><p>Requested source not defined</p>")
}

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

func stream(w http.ResponseWriter, url conf.Url) {
	f, err := getSocketFile(url.Source)
	if err != nil {
		return
	}
	c, err := net.FilePacketConn(f)
	if err != nil {
		log.Printf("Failed to get packet file connection: %s", err)
		return
	}
	f.Close()
	defer c.Close()
	host, _, err := net.SplitHostPort(url.Source)
	ipAddr := net.ParseIP(host)
	if err != nil {
		log.Printf("Cannot resolve address %s", url.Source)
		return
	}
	iface, _ := net.InterfaceByName(url.Interface)
	if err := ipv4.NewPacketConn(c).JoinGroup(iface, &net.UDPAddr{IP: net.IPv4(ipAddr[12], ipAddr[13], ipAddr[14], ipAddr[15])}); err != nil {
		log.Printf("Failed to join mulitcast group: %s", err)
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

func getSocketFile(address string) (*os.File, error) {
	host, port, err := net.SplitHostPort(address)
	ipAddr := net.ParseIP(host)
	dPort, _ := strconv.Atoi(port)
	s, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, 0)
	if err != nil {
		log.Printf("Syscall.Socket: %s", err)
		return nil, errors.New("Cannot create socket")
	}
	syscall.SetsockoptInt(s, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
	lsa := &syscall.SockaddrInet4{Port: dPort, Addr: [4]byte{ipAddr[12], ipAddr[13], ipAddr[14], ipAddr[15]}}
	if err := syscall.Bind(s, lsa); err != nil {
		log.Printf("Syscall.Bind: %s", err)
		return nil, errors.New("Cannot bind socket")
	}
	return os.NewFile(uintptr(s), "udp4:"+host+":"+port+"->"), nil
}

func loadUrlConfig() {
	c, err := conf.ReadUrls(urlsConfigPath)
	if err == nil {
		urls = c
	} else {
		log.Print("Config not loaded")
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.Parse()
	log.Printf("Process ID: %d", os.Getpid())
	loadUrlConfig()
	go osListener()
	http.HandleFunc("/", urlHandler)
	log.Printf("Listening on %s", *listenOn)
	http.ListenAndServe(*listenOn, nil)
}
