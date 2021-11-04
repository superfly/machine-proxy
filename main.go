package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/superfly/flyctl/api"
)

func copyHeader(dst, src http.Header) {
	for key, vals := range src {
		dst[key] = append([]string(nil), vals...)
	}
}

var client = http.Client{
	Timeout: time.Minute,
}

var apiClient *api.Client

type proxy struct {
	upstream string
}

var appName string

func (p *proxy) ServeHTTP(wr http.ResponseWriter, req *http.Request) {
	log.Printf("Incoming request %s %s %s", req.RemoteAddr, req.Method, req.URL)

	ctx, cancel := context.WithCancel(req.Context())
	defer cancel()

	started := make(chan bool)

	go startAppIfStopped(ctx, started)

	select {
	case <-ctx.Done():
		renderCode(wr, http.StatusGatewayTimeout) // this means we reached the timeout
		return
	case <-started:
		break // a healthy instance was found or one was booted; continue
	}

	fmt.Println("PROCESSING REQUEST")
	internalReq := req.Clone(req.Context())

	//http: Request.RequestURI can't be set in client requests.
	//http://golang.org/src/pkg/net/http/client.go
	internalReq.RequestURI = ""
	internalReq.URL.Host = p.upstream
	internalReq.URL.Scheme = "http"

	log.Printf("Internal request %s %s %s", internalReq.RemoteAddr, internalReq.Method, internalReq.URL)

	resp, err := client.Do(internalReq)
	if err != nil {
		renderCode(wr, http.StatusBadGateway)

		log.Printf("failed request: %v", err)

		return
	}
	defer resp.Body.Close()

	log.Println(internalReq.RemoteAddr, " ", resp.Status)

	copyHeader(wr.Header(), resp.Header)
	wr.WriteHeader(resp.StatusCode)
	io.Copy(wr, resp.Body)
}

func renderCode(w http.ResponseWriter, code int) {
	msg := http.StatusText(code)

	http.Error(w, msg, code)
}

func init() {
	log.SetOutput(os.Stderr)
	log.SetFlags(log.Ldate | log.Lmicroseconds)
}

func main() {

	ctx := context.Background()

	var addr = flag.String("addr", "127.0.0.1:8080", "The addr of the application.")
	flag.Parse()

	upstream, ok := os.LookupEnv("UPSTREAM")
	if !ok || upstream == "" {
		log.Fatal("UPSTREAM not defined")
	}

	appName, ok := os.LookupEnv("APP_NAME")
	if !ok || appName == "" {
		log.Fatal("APP_NAME not defined")
	}

	accessToken, ok := os.LookupEnv("FLY_ACCESS_TOKEN")
	if !ok || appName == "" {
		log.Fatal("FLY_ACCESS_TOKEN not defined")
	}

	api.SetBaseURL("https://app.fly.io")
	apiClient = api.NewClient(accessToken, "machines-proxy", "1.0.0", new(logger))

	log.Println("Starting app health check")
	go CheckAppHealth(ctx, appName, accessToken)

	handler := &proxy{
		upstream: upstream, // localhost:10201
	}

	log.Println("Starting proxy server on", *addr)
	if err := http.ListenAndServe(*addr, handler); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
