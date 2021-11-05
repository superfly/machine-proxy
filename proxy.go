package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"time"
)

var client = http.Client{
	Timeout: time.Minute,
}

type proxy struct {
	cfg      *config
	upstream string
}

func (p *proxy) do(w http.ResponseWriter, req *http.Request) {
	log.Println("proxying request")

	//http: Request.RequestURI can't be set in client requests.
	//http://golang.org/src/pkg/net/http/client.go

	req.RequestURI = ""
	req.URL.Host = p.upstream
	req.URL.Scheme = "http"

	log.Printf("request %s %s %s", req.RemoteAddr, req.Method, req.URL)

	res, err := client.Do(req)
	if err != nil {
		renderCode(w, http.StatusBadGateway)

		log.Printf("failed request: %v", err)

		return
	}
	defer res.Body.Close()

	log.Println(req.RemoteAddr, " ", res.Status)

	copyHeader(w.Header(), res.Header)
	w.WriteHeader(res.StatusCode)
	io.Copy(w, res.Body)
}

func (p *proxy) ServeHTTP(wr http.ResponseWriter, req *http.Request) {
	log.Printf("Incoming request %s %s %s", req.RemoteAddr, req.Method, req.URL)

	ctx, cancel := context.WithTimeout(req.Context(), time.Second*30)
	defer cancel()

	go ensureAppIsRunning(ctx, p.cfg)

	running, err := isAppRunning(ctx)
	switch {
	case err == nil:
		break
	case !running:
		renderCode(wr, http.StatusInternalServerError) // failed to run

		return
	default:
		renderCode(wr, http.StatusGatewayTimeout) // op timed out

		return
	}

	p.do(wr, req.Clone(ctx))
}

func renderCode(w http.ResponseWriter, code int) {
	msg := http.StatusText(code)

	http.Error(w, msg, code)
}

func copyHeader(dst, src http.Header) {
	for key, vals := range src {
		dst[key] = append([]string(nil), vals...)
	}
}
