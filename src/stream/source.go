package stream

import (
	"cast"
	"conf"
	"log"
)

type Source struct {
	Channel cast.Channel
	Config  conf.Url
}

// Create streamed sources
func Subscribe(config conf.UrlConfig) map[string]Source {
	sources := make(map[string]Source)
	for url, cfg := range config {
		scheme := conf.GetScheme(cfg.Source)
		if scheme == "udp" {
			cfg.Source = conf.GetHost(cfg.Source)
			sources[url] = Source{
				Channel: cast.New(100),
				Config: cfg,
			}
			log.Printf("Subcribing to %s source %s", scheme, sources[url].Config.Source)
			go UdpChannelStream(sources[url].Channel, sources[url].Config)
		} else if scheme == "http" {
			// TODO... or not
		}
	}
	return sources
}
