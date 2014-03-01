package conf

import (
	"log"
	"io/ioutil"
	"errors"
	"encoding/json"
	"net"
)

type Sets []uint

type RawNetworkConfig map[string]Sets

type NetworkConfigRecord struct {
	Network net.IPNet
	Sets    Sets
}

type NetworkConfig map[string]NetworkConfigRecord

// Read and parse JSON networks config
func ReadNetworks(fileName *string) (NetworkConfig, error) {
	log.Print("Reading config")
	file, err := ioutil.ReadFile(*fileName)
	if err != nil {
		log.Printf("Could not read config: %s", err)
		return NetworkConfig{}, errors.New("Could not read config")
	}
	var config RawNetworkConfig
	err = json.Unmarshal(file, &config)
	if err != nil {
		log.Printf("Could not parse config: %s", err)
		return NetworkConfig{}, errors.New("Could not parse config")
	}
	log.Printf("Read %d records", len(config))
	if cfg, result := networkConfigValid(config); result {
		return cfg, nil
	} else {
		return NetworkConfig{}, errors.New("Config is not valid")
	}
}

// Tells if networks config valid
func networkConfigValid(config RawNetworkConfig) (NetworkConfig, bool) {
	cfg := make(map[string]NetworkConfigRecord, len(config))
	for network, sets := range config {
		// Check for non-empty set
		if len(sets) == 0 {
			log.Printf("Empty set for network %s", network)
			return NetworkConfig{}, false
		}
		// Parse network address
		_, ipNet, err := net.ParseCIDR(network)
		if err != nil {
			log.Printf("Could not parse network address %s", network)
			return NetworkConfig{}, false
		}
		cfg[network] = NetworkConfigRecord{Network: *ipNet, Sets: sets}
	}
	return cfg, true
}
