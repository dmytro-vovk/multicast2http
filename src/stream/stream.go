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
)

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
	defer c.Close()
	host, _, err := net.SplitHostPort(url.Source)
	ipAddr := net.ParseIP(host)
	if err != nil {
		log.Printf("Cannot resolve address %s", url.Source)
		return nil, err
	}
	iface, _ := net.InterfaceByName(url.Interface)
	if err := ipv4.NewPacketConn(c).JoinGroup(iface, &net.UDPAddr{IP: net.IPv4(ipAddr[12], ipAddr[13], ipAddr[14], ipAddr[15])}); err != nil {
		log.Printf("Failed to join mulitcast group: %s", err)
		return nil, err
	}
	return c, nil
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
