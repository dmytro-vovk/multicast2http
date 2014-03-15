/**
 * @author Dmitry Vovk <dmitry.vovk@gmail.com>
 * @copyright 2014
 */
package stream

import (
	"syscall"
	"errors"
	"os"
	"log"
	"strconv"
	"net"
	"conf"
	"code.google.com/p/go.net/ipv4"
	"io"
	"response"
	"net/http"
	"cast"
)

// Stream subscribed channel
func ChannelStream(w http.ResponseWriter, src Source) {
	ch := make(chan interface{})
	src.Channel.Join(ch)
	defer src.Channel.Leave(ch)
	for {
		payload := <-ch
		var b []byte
		b = payload.([]byte)
		if _, err := w.Write(b); err != nil {
			return
		}
	}
}

// Broadcaster for UDP stream
func UdpChannelStream(ch cast.Channel, url conf.Url) {
	c, err := GetStreamSource(url)
	if err != nil {
		return
	}
	defer c.Close()
	b := make([]byte, conf.MaxMTU)
	localAddress := c.LocalAddr().String()
	for {
		n, _, err := c.ReadFrom(b)
		if err != nil {
			log.Printf("Failed to read from UDP stream %s: %s", url.Source, err)
			return
		}
		if url.Source == localAddress {
			ch.Send(b[:n]);
		}
	}
}

// Perform actual unicast streaming
func HttpStream(w http.ResponseWriter, url conf.Url) {
	r, err := http.Get(url.Source)
	if err != nil {
		log.Printf("Failed to open HTTP stream %s: %s", url.Source, err)
		response.NotFound(w)
		return
	}
	defer r.Body.Close()
	io.Copy(w, r.Body)
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
