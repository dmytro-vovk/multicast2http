package server

import (
	"conf"
	"hls"
	"log"
	"net/http"
	"os"
	"os/signal"
	"response"
	"syscall"
)

func init() {
	go osListener()
	go hupListener()
}

// OS stop signals listener
func osListener() {
	osExitSignals := make(chan os.Signal, 1)
	signal.Notify(osExitSignals, os.Interrupt, os.Kill)
	signal := <-osExitSignals
	hls.StopStreams()
	log.Fatalf("Exiting due to %s", signal)
}

// OS HUP signal listener
func hupListener() {
	osHupSignals := make(chan os.Signal, 1)
	signal.Notify(osHupSignals, syscall.SIGHUP)
	for {
		<-osHupSignals
		conf.RereadConfigs()
	}
}

func ReloadConfigs(w http.ResponseWriter, _ *http.Request) {
	conf.RereadConfigs()
	response.ConfigReloaded(w)
}
