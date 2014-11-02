package hlstream

import (
	"code.google.com/p/go.net/ipv4"
	"conf"
	"errors"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

var (
	HLSDir, Ffmpeg string
)

const (
	OPTIONS_FFMPEG = "-y -i - -threads 4 "
	//OPTIONS_AUDIO  = "-c:a aac -ac 2 -strict -2 "
	OPTIONS_AUDIO  = "-acodec copy "
	//OPTIONS_VIDEO  = "-c:v libx264 -vprofile baseline -x264opts level=41 "
	OPTIONS_VIDEO  = "-vcodec copy "
	OPTIONS_HLS    = "-hls_time 2 -hls_list_size 5 -hls_wrap 10 -start_number 1 -re -segment_list_flags +live "
	PLAYLIST_FILE  = "/stream.m3u8"
)

func RunStreams(config conf.UrlConfig) {
	for url, cfg := range config {
		go streamer(url, cfg)
	}
}

// Run single ffmpeg HLS stream
func streamer(url string, cfg conf.Url) {
	// Prepare output dir
	destinationDir := getDir(url)
	// Prepare ffmpeg
	cmd := exec.Command(
		Ffmpeg,
		strings.Split(OPTIONS_FFMPEG+OPTIONS_AUDIO+OPTIONS_VIDEO+OPTIONS_HLS+destinationDir+PLAYLIST_FILE, " ")...,
	)
	feed, _ := cmd.StdinPipe()
	out, _ := cmd.StderrPipe()
	err := cmd.Start()
	if err != nil {
		errOut, _ := ioutil.ReadAll(out)
		log.Printf("Ffmpeg start: %s:\n%s", err, string(errOut))
	}
	go func() {
		err := cmd.Wait()
		if err != nil {
			errOut, _ := ioutil.ReadAll(out)
			log.Printf("Ffmpeg stop: %s:\n%s", err, string(errOut))
		}
	}()
	// Prepare the stream
	c, err := GetStreamSource(cfg)
	if err != nil {
		log.Printf("Failed to get stream %s source: %s", cfg.Source, err)
		return
	}
	defer c.Close()
	src, _, _ := net.SplitHostPort(cfg.Source)
	localAddress := c.LocalAddr().String()
	h, _, _ := net.SplitHostPort(localAddress)
	if src == h {
		b := make([]byte, conf.MaxMTU)
		for {
			n, _, err := c.ReadFrom(b)
			if err != nil {
				log.Printf("Failed to read from UDP stream %s: %s", src, err)
				return
			}
			_, err = feed.Write(b[:n])
			if err != nil {
				log.Printf("Got error during write: %s", err)
				return
			}
		}
	}
}

// Prepare full path for give URL and make sure dir exists
func getDir(url string) string {
	destinationDir := HLSDir + url
	if _, err := os.Stat(destinationDir); os.IsNotExist(err) {
		if err = os.MkdirAll(destinationDir, 0777); err != nil {
			log.Fatalf("Could not create stream directory %s", destinationDir)
		}
	}
	return destinationDir
}

// Returns UDP Multicast packet connection to read incoming bytes from
func GetStreamSource(url conf.Url) (net.PacketConn, error) {
	f, err := getSocketFile(url.Source)
	if err != nil {
		return nil, err
	}
	c, err := net.FilePacketConn(f)
	if err != nil {
		log.Printf("Failed to get packet file connection: %s", err)
		return nil, err
	}
	f.Close()
	host, _, err := net.SplitHostPort(url.Source)
	ipAddr := net.ParseIP(host).To4()
	if err != nil {
		log.Printf("Cannot resolve address %s", url.Source)
		return nil, err
	}
	iface, _ := net.InterfaceByName(url.Interface)
	if err := ipv4.NewPacketConn(c).JoinGroup(iface, &net.UDPAddr{IP: net.IPv4(ipAddr[0], ipAddr[1], ipAddr[2], ipAddr[3])}); err != nil {
		log.Printf("Failed to join mulitcast group: %s", err)
		return nil, err
	}
	return c, nil
}

// Returns bound UDP socket
func getSocketFile(address string) (*os.File, error) {
	host, port, err := net.SplitHostPort(address)
	ipAddr := net.ParseIP(host).To4()
	dPort, _ := strconv.Atoi(port)
	s, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, 0)
	if err != nil {
		log.Printf("Syscall.Socket: %s", err)
		return nil, errors.New("Cannot create socket")
	}
	syscall.SetsockoptInt(s, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
	lsa := &syscall.SockaddrInet4{Port: dPort, Addr: [4]byte{ipAddr[0], ipAddr[1], ipAddr[2], ipAddr[3]}}
	if err := syscall.Bind(s, lsa); err != nil {
		log.Printf("Syscall.Bind: %s", err)
		return nil, errors.New("Cannot bind socket")
	}
	return os.NewFile(uintptr(s), "udp4:"+host+":"+port+"->"), nil
}
