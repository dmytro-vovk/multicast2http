/**
 * @author Dmitry Vovk <dmitry.vovk@gmail.com>
 * @copyright 2014
 */
package conf

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net"
	netUrl "net/url"
	"regexp"
	"strconv"
)

type Url struct {
	Source      string      `json:"source"`      // IP:port of stream source
	FfmpegArgs  string      `json:"ffmpeg-args"` // ...or string or arguments to pass to ffmpeg
	Interface   string      `json:"interface"`   // NIC name
	Deinterlace bool        `json:"deinterlace"` // Whether to deinterlace (valid only for HLS)
	CopyStream  bool        `json:"copy-stream"` // Do not transcode (valid only for HLS)
	Set         uint        `json:"set"`         // Set id
	Networks    []net.IPNet `json:"-"`           // Allowed networks (based on set values in networks config)
}

type webConfig struct {
	EnableControls bool   `json:"enable-controls"`
	AllowOrigin    string `json:"allow-origin"`
}

type hlsConfig struct {
	Dir      string `json:"dir"`
	Ffmpeg   string `json:"ffmpeg"`
	ChunkLen int    `json:"chunk-len"`
}

type configType struct {
	Sources    string    `json:"sources"`
	Networks   string    `json:"networks"`
	Listen     string    `json:"listen"`
	FakeStream string    `json:"fake-stream"`
	Web        webConfig `json:"web"`
	Hls        hlsConfig `json:"hls"`
	Urls       UrlConfig
}

type UrlConfig map[string]Url

var (
	MaxMTU         int = 1500
	config         configType
	configFileName string
)

const (
	VALID_PATH = `^/[a-z0-9_-]+$`
)

func Conf() configType {
	return config
}

// Read and parse JSON sources config
func ReadUrls(fileName string) (UrlConfig, error) {
	log.Print("Reading sources from " + fileName)
	file, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Printf("Could not read sources: %s", err)
		return UrlConfig{}, errors.New("Could not read config")
	}
	var config UrlConfig
	err = json.Unmarshal(file, &config)
	if err != nil {
		log.Printf("Could not parse sources: %s", err)
		return UrlConfig{}, errors.New("Could not parse config")
	}
	log.Printf("Found %d records", len(config))
	if configValid(config) {
		return config, nil
	} else {
		return UrlConfig{}, errors.New("Config is not valid")
	}
}

// Check config for validity
func configValid(config UrlConfig) bool {
	// Get list of network interfaces
	var ifaceNames map[string]bool
	interfaces, err := net.Interfaces()
	if err != nil {
		log.Printf("Could not get interfaces list: %s", err)
		return false
	}
	ifaceNames = make(map[string]bool, len(interfaces))
	for _, iface := range interfaces {
		ifaceNames[iface.Name] = true
		// Find out biggest MTU
		if iface.MTU > MaxMTU {
			MaxMTU = iface.MTU
		}
	}
	var validPath = regexp.MustCompile(VALID_PATH)
	for path, url := range config {
		if !validPath.MatchString(path) {
			log.Printf("Invalid path found: %s", path)
			return false
		}
		if url.FfmpegArgs != "" && url.Source != "" {
			log.Printf("In source %s use either source or ffmpeg-args, but not both")
			return false
		}
		if url.FfmpegArgs == "" {
			log.Printf("Parsing %s", url.Source)
			host, port, err := net.SplitHostPort(url.Source)
			if err != nil {
				hostUrl, err := netUrl.Parse(url.Source)
				var rawHost string
				if err != nil {
					log.Printf("Could not parse source ip:port: %s", rawHost)
					return false
				}
				host, port, err = net.SplitHostPort(hostUrl.Host)
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
	}
	return true
}

func RereadConfigs() {
	LoadConfig(configFileName)
}

// Reread sources config
func LoadConfig(configFile string) {
	var conf configType
	log.Print("Reading config " + configFile)
	file, err := ioutil.ReadFile(configFile)
	if err == nil {
		err = json.Unmarshal(file, &conf)
	}
	if err != nil {
		log.Fatal(err)
	}
	configFileName = configFile
	config = conf
	_urls, err := ReadUrls(config.Sources)
	if err == nil {
		log.Print("Reading networks from " + config.Networks)
		_nets, err := ReadNetworks(config.Networks)
		if err == nil {
			config.Urls = mergeConfigs(_urls, _nets)
		} else {
			log.Print("Networks not loaded")
		}
	} else {
		log.Print("Sources not loaded")
	}
}

// Populate sources with allowed networks based on sets
func mergeConfigs(_urls UrlConfig, _nets NetworkConfig) UrlConfig {
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
