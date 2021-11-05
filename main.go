package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/azazeal/pause"
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
	cfg, err := configFromEnv()
	if err != nil {
		log.Fatalf("failed loading config: %v", err)
	}

	api.SetBaseURL("https://app.fly.io")
	apiClient = api.NewClient(cfg.accessToken, "machines-proxy", "1.0.0", new(logger))

	run(cfg)
}

func run(cfg *config) {
	var wg sync.WaitGroup
	defer wg.Wait()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	go func() {
		defer wg.Done()

		checkAppHealth(ctx, cfg)
	}()

	go func() {
		defer wg.Done()

		serve(ctx, cfg.addr, cfg.upstream)
	}()
}

func serve(ctx context.Context, addr, upstream string) {
	log.Println("entered serve")
	defer log.Println("exited serve")

	handler := &proxy{
		upstream: upstream, // localhost:10201
	}

	loop(ctx, time.Second, func(context.Context) {
		if err := http.ListenAndServe(addr, handler); err != http.ErrServerClosed {
			log.Printf("failed listening: %v", err)
		}
	})
}

func loop(ctx context.Context, p time.Duration, fn func(context.Context)) {
	for ctx.Err() == nil {
		fn(ctx)

		pause.For(ctx, p)
	}
}
