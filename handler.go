package hsts

import (
	"net/http"
	"strconv"
	"time"
)

// Handler is a HTTP handler which will redirect any request served over HTTP over to HTTPS and apply a
// HSTS header.
type Handler struct {
	next http.Handler
	// MaxAge sets the duration that the HSTS is valid for.
	MaxAge time.Duration
	// HostOverride provides a host to the redirection URL in the case that the system is behind a load balancer
	// which doesn't provide the X-Forwarded-Host HTTP header (e.g. an Amazon ELB).
	HostOverride string
	// Decides whether to accept the X-Forwarded-Proto header as proof of SSL.
	AcceptXForwardedProtoHeader bool
}

// NewHandler creates a new HSTS redirector, which will redirect any request served over HTTP over to HTTPS.
func NewHandler(next http.Handler) *Handler {
	return &Handler{
		next:   next,
		MaxAge: time.Hour * 24 * 90, // 90 days
		AcceptXForwardedProtoHeader: true,
	}
}

func isHTTPS(r *http.Request, acceptXForwardedProtoHeader bool) bool {
	// Added by common load balancers which do SSL offloading.
	if acceptXForwardedProtoHeader && r.Header.Get("X-Forwarded-Proto") == "https" {
		return true
	}
	// Set by some middleware.
	if r.URL.Scheme == "https" {
		return true
	}
	// Set when the Go server is running HTTPS itself.
	if r.TLS != nil && r.TLS.HandshakeComplete {
		return true
	}

	return false
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if isHTTPS(r, h.AcceptXForwardedProtoHeader) {
		timeInSeconds := int(h.MaxAge.Seconds())
		w.Header().Add("Strict-Transport-Security", "max-age="+strconv.Itoa(timeInSeconds)+"; includeSubDomains")

		h.next.ServeHTTP(w, r)
	} else {
		if h.HostOverride != "" {
			r.URL.Host = h.HostOverride
		} else if !r.URL.IsAbs() {
			r.URL.Host = r.Host
		}

		r.URL.Scheme = "https"

		http.Redirect(w, r, r.URL.String(), http.StatusMovedPermanently)
	}
}
