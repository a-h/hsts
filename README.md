# hsts
Go (Golang) middleware which redirects users from HTTP to HTTPS and adds 
the HSTS header.

# Usage

```go
package main

import (
	"log"
	"net/http"

	"github.com/a-h/hsts"
)

func main() {
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
The example application demonstrates how the application redirects the user and adds the HSTS header if HTTP is used.

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
< Strict-Transport-Security: max-age=15768000; includeSubDomains
< Date: Fri, 29 Jul 2016 12:18:37 GMT
< Content-Length: 70
< Content-Type: text/html; charset=utf-8
<
<a href="https://localhost:8080/hellohandler">Moved Permanently</a>.

* Connection #0 to host localhost left intact
```
