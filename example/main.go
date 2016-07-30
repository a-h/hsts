package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/a-h/hsts"
)

var serveHTTPSFlag = flag.Bool("serveHTTPS", false, "Whether the system should serve SSL using the example certificate.")

func main() {
	if *serveHTTPSFlag {
		serveHTTPS()
	} else {
		serveHTTP()
	}
}

func serveHTTPS() {
	log.Print("Serving HTTPS to demonstrate header.")

	http.Handle("/hellohandler", hsts.NewHandler(&structHandler{}))
	http.Handle("/hellofunction", hsts.NewHandler(http.HandlerFunc(serveFunction)))
	log.Fatal(http.ListenAndServeTLS(":443", "server.pem", "server.key", nil))
}

func serveHTTP() {
	redirectToSSL := true

	if redirectToSSL {
		log.Print("Serving with HTTP to HTTPS redirection.")

		http.Handle("/hellohandler", hsts.NewHandler(&structHandler{}))
		http.Handle("/hellofunction", hsts.NewHandler(http.HandlerFunc(serveFunction)))
		log.Fatal(http.ListenAndServe(":8080", nil))
	} else {
		log.Print("Serving without HTTP to HTTPS redirection.")

		http.Handle("/hellohandler", &structHandler{})
		http.HandleFunc("/hellofunction", serveFunction)
		log.Fatal(http.ListenAndServe(":8080", nil))
	}
}

type structHandler struct {
}

func (h *structHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello from the Handler interface."))
}

func serveFunction(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello from a function."))
}
