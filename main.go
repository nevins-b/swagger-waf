package main

import (
	"flag"
	"log"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/nevins-b/gWAF/waf"
)

func main() {
	var listen, swagger, backend string
	var logOnly, verbose bool
	flag.StringVar(&listen, "listen", ":8000", "Address to listen for requests on")
	flag.StringVar(&swagger, "swagger", "", "Path to the swagger spec")
	flag.StringVar(&backend, "backend", "", "Address to proxy requests too")
	flag.BoolVar(&logOnly, "logonly", false, "Only log, do not block")
	flag.BoolVar(&verbose, "verbose", false, "Verbose logging")
	flag.Parse()

	ctx := waf.InitContext(listen, backend, swagger, logOnly, verbose)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[%v] %s %s", time.Now(), r.Method, r.URL.RequestURI())
		if !waf.Valid(ctx, r) {
			w.WriteHeader(500)
			return
		}

		// Everything looks good, proxy to backend
		director := func(req *http.Request) {
			req = r
			req.URL.Scheme = "http"
			req.URL.Host = ctx.Backend
		}
		proxy := &httputil.ReverseProxy{Director: director}
		proxy.ServeHTTP(w, r)
	})

	s := &http.Server{
		Addr:           listen,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	log.Fatal(s.ListenAndServe())
}
