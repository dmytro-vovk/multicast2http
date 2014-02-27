/**
 * @author Dmitry Vovk <dmitry.vovk@gmail.com>
 * @copyright 2014
 */
package conf

import (
	"log"
	"io/ioutil"
	"encoding/json"
	"errors"
	"regexp"
	"net"
	"strconv"
)

type Url struct {
	Source    string `json:"source"`
	Interface string `json:"interface"`
	Set       uint `json:"set"`
}

type UrlConfig map[string]Url

const (
	VALID_PATH = `^/[a-z0-9_-]+$`
)

/**
 * Read and parse JSON config
 */
func ReadUrls(fileName *string) (UrlConfig, error) {
	log.Print("Reading config")
	file, err := ioutil.ReadFile(*fileName)
	if err != nil {
		log.Printf("Could not read config: %s", err)
		return UrlConfig{}, errors.New("Could not read config")
	}
	var config UrlConfig
	err = json.Unmarshal(file, &config)
	if err != nil {
		log.Printf("Could not parse config: %s", err)
		return UrlConfig{}, errors.New("Could not parse config")
	}
	log.Printf("Read %d records", len(config))
	if configValid(config) {
		return config, nil
	} else {
		return UrlConfig{}, errors.New("Config is not valid")
	}
}

/**
 * Check config for validity:
 * - urls match regexp pattern
 * - ip:port pairs are valid ip and port number
 * - interfaces exists
 */
func configValid(config UrlConfig) bool {
	var validPath = regexp.MustCompile(VALID_PATH)
	var ifaceNames map[string]bool
	interfaces, err := net.Interfaces()
	if err != nil {
		log.Printf("Could not get interfaces list: %s", err)
		return false
	}
	ifaceNames = make(map[string]bool, len(interfaces))
	for _, iface := range interfaces {
		ifaceNames[iface.Name] = true
	}
	for path, url := range config {
		if !validPath.MatchString(path) {
			log.Printf("Invalid path found: %s", path)
			return false
		}
		host, port, err := net.SplitHostPort(url.Source)
		if err != nil {
			log.Printf("Could not parse host:port source pair of path %s: %s", path, url.Source)
			return false
		}
		ipAddr := net.ParseIP(host)
		if ipAddr == nil {
			log.Printf("Invalid ip address in source %s: %s", path, host)
			return false
		}
		dPort, err := strconv.Atoi(port)
		if dPort == 0 || err != nil {
			log.Printf("Invalid port in source %s: %s", path, port)
			return false
		}
		if _, ok := ifaceNames[url.Interface]; !ok {
			log.Printf("Interface for source %s not found: %s", path, url.Interface)
			return false
		}
	}
	return true
}
