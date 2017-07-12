package hsts

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestThatHTTPRequestsAreRedirectedToHTTPS(t *testing.T) {
	// Create the handler to wrap.
	wrappedHandler := &TestHandler{
		body: []byte("Shouldn't see this content over HTTP, only over HTTPS."),
	}

	// Create the middleware and pass it the wrapped handler.
	hstsHandler := NewHandler(wrappedHandler)

	redirectionTests := []struct {
		inputURL         string
		expectedRedirect string
		method           string
	}{
		{
			inputURL:         "http://www.example.com",
			expectedRedirect: "https://www.example.com",
			method:           http.MethodGet,
		},
		{
			inputURL:         "http://example.com",
			expectedRedirect: "https://example.com",
			method:           http.MethodPost,
		},
		{
			inputURL:         "http://example.com:80",
			expectedRedirect: "https://example.com:80",
			method:           http.MethodPost,
		},
		{
			inputURL:         "http://example.com/test/",
			expectedRedirect: "https://example.com/test/",
			method:           http.MethodPost,
		},
		{
			inputURL:         "/test",
			expectedRedirect: "https://example.com/test",
			method:           http.MethodPost,
		},
	}

	for _, test := range redirectionTests {
		// Create a mock request to capture the result.
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(test.method, test.inputURL, nil)
		req.Host = "example.com"

		// Act.
		hstsHandler.ServeHTTP(w, req)

		// Assert.
		actual := struct {
			body         string
			redirectedTo string
			statusCode   int
			stsHeader    string
		}{
			body:         strings.TrimSpace(string(w.Body.Bytes())),
			redirectedTo: w.Header().Get("Location"),
			statusCode:   w.Code,
			stsHeader:    w.Header().Get("Strict-Transport-Security"),
		}

		if test.expectedRedirect != actual.redirectedTo {
			t.Errorf("For a HTTP %s to %s, expected redirect to %s, but was %s", test.method, test.inputURL, test.expectedRedirect, actual.redirectedTo)
		}

		if http.StatusMovedPermanently != actual.statusCode {
			t.Errorf("For a HTTP %s to %s, expected a status code of 301 Moved Permanently but was %d", test.method, test.inputURL, actual.statusCode)
		}

		if test.method == http.MethodGet {
			expectedBody := "<a href=\"" + test.expectedRedirect + "\">Moved Permanently</a>."

			if actual.body != expectedBody {
				t.Errorf("For a HTTP %s to %s, expected a redirect link in the body \"%s\" but the body contained \"%s\"", test.method, test.inputURL, expectedBody, actual.body)
			}
		}

		if actual.stsHeader != "" {
			t.Errorf("The STS header should only be set on HTTPS requests. But over HTTP it was %s", actual.stsHeader)
		}
	}
}

func TestThatCommonHTTPHeadersAreUsedToDetermineSSLStatus(t *testing.T) {
	r, _ := http.NewRequest("GET", "/", nil)

	if isHTTPS(r, true) {
		t.Error("The request was HTTP, but this was not detected")
	}

	r.Header.Add("X-Forwarded-Proto", "https")

	if !isHTTPS(r, true) {
		t.Error("The request was HTTPS, but this was not detected")
	}

	if isHTTPS(r, false) {
		t.Error("The request was declared HTTPS by the header, but this was not enabled.")
	}
}

func TestThatTLSIsUsedToDetermineSSLStatus(t *testing.T) {
	r, _ := http.NewRequest("GET", "/", nil)
	r.TLS = &tls.ConnectionState{
		HandshakeComplete: true,
	}

	if !isHTTPS(r, false) {
		t.Error("The request was HTTPS, but this was not detected")
	}
}

func TestThatTLSDoesNotOverrideHTTPHeadersToDetermineSSLStatus(t *testing.T) {
	r, _ := http.NewRequest("GET", "/", nil)
	r.TLS = &tls.ConnectionState{
		HandshakeComplete: true,
	}
	r.Header.Add("X-Forwarded-Proto", "http")

	if isHTTPS(r, true) {
		t.Error("The request was HTTPS but was declared HTTP by the header, but this was overruled")
	}
}

func TestThatTheHostCanBeOverridden(t *testing.T) {
	// Create the handler to wrap.
	wrappedHandler := &TestHandler{
		body: []byte("Shouldn't see this content over HTTP, only over HTTPS."),
	}

	// Create the middleware and pass it the wrapped handler.
	hstsHandler := NewHandler(wrappedHandler)
	hstsHandler.HostOverride = "subdomain.example.com"

	// Create a mock request to capture the result.
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "http://example.com/test", nil)

	// Act.
	hstsHandler.ServeHTTP(w, req)

	// Assert.
	expectedRedirect := "https://subdomain.example.com/test"
	actualRedirect := w.Header().Get("Location")
	if actualRedirect != expectedRedirect {
		t.Errorf("Expected a redirect to %s, but got a redirect to %s", expectedRedirect, actualRedirect)
	}
}

func TestThatHTTPSRequestsHaveTheHeaderApplied(t *testing.T) {
	// Create the handler to wrap.
	wrappedHandler := &TestHandler{
		body: []byte("Secure content!"),
	}

	// Create the middleware and pass it the wrapped handler.
	hstsHandler := NewHandler(wrappedHandler)

	urls := []string{
		"https://www.example.com",
		"https://example.com",
		"https://example.com:443",
		"https://example.com:443/test/another?query=123",
		"https://example.com:443/test/another?query=123&more=456#789",
	}

	for _, url := range urls {
		// Create a mock request to capture the result.
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", url, nil)

		// Act.
		hstsHandler.ServeHTTP(w, req)

		// Assert.
		actual := struct {
			body         string
			redirectedTo string
			statusCode   int
			stsHeader    string
		}{
			body:         string(w.Body.Bytes()),
			redirectedTo: w.Header().Get("Location"),
			statusCode:   w.Code,
			stsHeader:    w.Header().Get("Strict-Transport-Security"),
		}

		if actual.redirectedTo != "" {
			t.Errorf("For a HTTPS GET to %s, expected no redirect but received one to %s", url, actual.redirectedTo)
		}

		if actual.statusCode != http.StatusOK {
			t.Errorf("For a HTTPS GET to %s, expected a status code of 200 OK but was %d", url, actual.statusCode)
		}

		if actual.body != "Secure content!" {
			t.Errorf("For a HTTPS GET to %s, expected a body of %s but the body was actually %s", url, "Secure content!", actual.body)
		}

		if !strings.Contains(actual.stsHeader, "max-age=10886400;") {
			t.Errorf("Expected the STS header to contain a max-age of 10886400 (18 weeks in seconds), but the header was \"%s\"", actual.stsHeader)
		}
	}
}

func TestHeaderValueIncludesPreloadDirective(t *testing.T) {
	expected := "max-age=3600; includeSubDomains"
	actual := createHeaderValue(time.Hour, false)

	if expected != actual {
		t.Errorf("Expected header without preload of \"%s\", but got \"%s\"", expected, actual)
	}

	expected = "max-age=3600; includeSubDomains; preload"
	actual = createHeaderValue(time.Hour, true)

	if expected != actual {
		t.Errorf("Expected header with a preload of \"%s\", but got \"%s\"", expected, actual)
	}
}

type TestHandler struct {
	body        []byte
	wasExecuted bool
}

func (th *TestHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.Write(th.body)
	th.wasExecuted = true
	w.Header().Add("x-test", "x-test-value")
}

func BenchmarkCreateheaderValueOld(b *testing.B) {
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		createHeaderValue(time.Hour, true)
	}
}

func BenchmarkCreateheaderValueNew(b *testing.B) {
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		createHeaderValueNew(time.Hour, true)
	}
}
