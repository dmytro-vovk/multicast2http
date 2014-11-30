package response

import (
	"conf"
	"fmt"
	"net/http"
	"time"
)

var (
	startTime = time.Now()
)

/**
 * 404 Not Found page
 */
func NotFound(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(404)
	fmt.Fprintln(w, "<h1>Not found</h1><p>Requested source not defined</p>")
}

/**
 * 500 Internal Error page
 */
func ServerFail(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(500)
	fmt.Fprintln(w, "<h1>Internal Error</h1><p>"+message+"</p>")
}

/**
 * Responds with server status page
 */
func ShowStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintln(w, "<h1>Server Status</h1>")
	fmt.Fprintf(w, "<p>Uptime: %s</p>", time.Since(startTime))
	fmt.Fprintf(w, "<p>Clients: %d</p>", conf.Stats.RunningStreams)
}

func ConfigReloaded(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintln(w, "<h1>Configs Reloaded</h1>")
}
