package hsts

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestThatHTTPRequestsAreRedirectToHTTPS(t *testing.T) {
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
			inputURL:         "http://www.google.com",
			expectedRedirect: "https://www.google.com",
			method:           http.MethodGet,
		},
		{
			inputURL:         "http://www.google.com",
			expectedRedirect: "https://www.google.com",
			method:           http.MethodPost,
		},
	}

	for _, test := range redirectionTests {
		// Create a mock request to capture the result.
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(test.method, test.inputURL, nil)

		// Act.
		hstsHandler.ServeHTTP(w, req)

		// Assert.
		actual := struct {
			body         string
			redirectedTo string
			statusCode   int
		}{
			body:         strings.TrimSpace(string(w.Body.Bytes())),
			redirectedTo: w.Header().Get("Location"),
			statusCode:   w.Code,
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
	}
}

func TestThatHTTPSRequestsAreNotAffected(t *testing.T) {
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
		}{
			body:         string(w.Body.Bytes()),
			redirectedTo: w.Header().Get("Location"),
			statusCode:   w.Code,
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
