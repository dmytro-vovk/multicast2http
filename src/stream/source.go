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
		cfg.Source = conf.GetHost(cfg.Source)
		sources[url] = Source{
			Channel: cast.New(1),
			Config: cfg,
		}
		log.Printf("Subcribing to %s source %s", scheme, sources[url].Config.Source)
		if scheme == "udp" {
			go streamUdpSource(sources[url])
		} else if scheme == "http" {
			// TODO
		}
	}
	return sources
}

func streamUdpSource(src Source) {
	UdpChannelStream(src.Channel, src.Config)
}
