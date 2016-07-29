package hsts

import "net/http"

// Handler is a HTTP handler which will redirect any request served over HTTP over to HTTPS and apply a
// HSTS header.
type Handler struct {
	next http.Handler
}

// NewHandler creates a new HSTS redirector, which will redirect any request served over HTTP over to HTTPS.
func NewHandler(next http.Handler) *Handler {
	return &Handler{
		next: next,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Scheme == "http" {
		w.Header().Add("Strict-Transport-Security", "max-age=15768000; includeSubDomains")

		r.URL.Scheme = "https"
		http.Redirect(w, r, r.URL.String(), http.StatusMovedPermanently)
	} else {
		h.next.ServeHTTP(w, r)
	}
}
