package hsts

import (
	"net/http"
	"strconv"
	"time"
)

// Handler is a HTTP handler which will redirect any request served over HTTP over to HTTPS and apply a
// HSTS header.
type Handler struct {
	next   http.Handler
	maxAge time.Duration
}

// NewHandler creates a new HSTS redirector, which will redirect any request served over HTTP over to HTTPS.
func NewHandler(next http.Handler) *Handler {
	return &Handler{
		next:   next,
		maxAge: time.Hour * 24 * 90, // 90 days
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Scheme == "http" || r.URL.Scheme == "" {
		timeInSeconds := int(h.maxAge.Seconds())
		w.Header().Add("Strict-Transport-Security", "max-age="+strconv.Itoa(timeInSeconds)+"; includeSubDomains")

		if !r.URL.IsAbs() {
			r.URL.Host = r.Host
		}

		r.URL.Scheme = "https"

		http.Redirect(w, r, r.URL.String(), http.StatusMovedPermanently)
	} else {
		h.next.ServeHTTP(w, r)
	}
}
