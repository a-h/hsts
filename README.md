# hsts
Go (Golang) middleware which redirects users from HTTP to HTTPS and adds 
the HSTS header.

# Usage

```go
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
```

# Example
The example application demonstrates how the application redirects the user.

```bash
cd example
go run main.go
```

```bash
curl -v localhost:8080/hellohandler
```

```bash
*   Trying ::1...
* Connected to localhost (::1) port 8080 (#0)
> GET /hellohandler HTTP/1.1
> Host: localhost:8080
> User-Agent: curl/7.43.0
> Accept: */*
>
< HTTP/1.1 301 Moved Permanently
< Location: https://localhost:8080/hellohandler
< Date: Sat, 30 Jul 2016 09:58:06 GMT
< Content-Length: 70
< Content-Type: text/html; charset=utf-8
<
<a href="https://localhost:8080/hellohandler">Moved Permanently</a>.

* Connection #0 to host localhost left intact
```

# Example (SSL)
This example demonstrates adding the HSTS header when content is served over HTTPS.

This example has to be compiled and ran as sudo in order to be able to use port 443 for SSL. 

```bash
sudo ./example -serveHTTPS=true
```

```bash
curl -k https://localhost/hellohandler -v
```

```bash
*   Trying ::1...
* Connected to localhost (::1) port 443 (#0)
* TLS 1.2 connection using TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256
* Server certificate: Internet Widgits Pty Ltd
> GET /hellohandler HTTP/1.1
> Host: localhost
> User-Agent: curl/7.43.0
> Accept: */*
>
< HTTP/1.1 200 OK
< Strict-Transport-Security: max-age=7776000; includeSubDomains
< Date: Sat, 30 Jul 2016 10:34:53 GMT
< Content-Length: 33
< Content-Type: text/plain; charset=utf-8
<
* Connection #0 to host localhost left intact
Hello from the Handler interface.
```