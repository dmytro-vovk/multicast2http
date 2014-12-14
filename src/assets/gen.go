//go:generate /usr/local/golib/bin/go-bindata -pkg=assets public/
package assets

import "net/http"

func HandlerHtml(w http.ResponseWriter, _ *http.Request) {
	b, _ := Asset("public/index.html")
	w.Header().Set("Content-Type", "text/html")
	w.Write(b)
}

func HandlerJs(w http.ResponseWriter, _ *http.Request) {
	b, _ := Asset("public/player.js")
	w.Header().Set("Content-Type", "application/javascript")
	w.Write(b)
}

func HandlerSwf(w http.ResponseWriter, _ *http.Request) {
	b, _ := Asset("public/player.swf")
	w.Header().Set("Content-Type", "application/x-shockwave-flash")
	w.Write(b)
}
