package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/azazeal/pause"
	"github.com/superfly/flyctl/api"
)

func init() {
	log.SetOutput(os.Stderr)
	log.SetFlags(log.Ldate | log.Lmicroseconds)
}

func main() {
	api.SetBaseURL("https://app.fly.io")

	cfg, err := configFromEnv()
	if err != nil {
		log.Fatalf("failed loading config: %v", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(),
		os.Interrupt, syscall.SIGTERM)
	defer cancel()

	run(ctx, cfg)
}

func run(ctx context.Context, cfg *config) {
	var wg sync.WaitGroup
	defer wg.Wait()

	wg.Add(1)
	go func() {
		defer wg.Done()

		checkAppHealth(ctx, cfg)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		serve(ctx, cfg)
	}()
}

func serve(ctx context.Context, cfg *config) {
	log.Println("entered serve")
	defer log.Println("exited serve")

	handler := &proxy{
		upstream: cfg.upstream, // localhost:10201
	}

	loop(ctx, time.Second, func(context.Context) {
		l, err := net.Listen("tcp", cfg.addr)
		if err != nil {
			log.Printf("failed listening on %s: %v", cfg.addr, err)

			return
		}

		srv := &http.Server{
			Handler:  handler,
			ErrorLog: log.New(log.Writer(), log.Prefix(), log.Flags()),
		}

		served := make(chan struct{})
		defer close(served)

		go func() {
			select {
			case <-ctx.Done():
				ctx, cancel := context.WithTimeout(context.Background(), time.Minute>>1)
				defer cancel()

				_ = srv.Shutdown(ctx)
			case <-served:
				break
			}
		}()

		if err := srv.Serve(l); err != http.ErrServerClosed {
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
