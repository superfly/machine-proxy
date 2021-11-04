package main

import (
	"context"
	"sync"
	"time"

	"github.com/azazeal/pause"
)

var (
	appHealthyMutex sync.RWMutex
	appHealthy      bool
)

func CheckAppHealth(ctx context.Context, appName string, accessToken string) {
	loop(ctx, time.Second, func(ctx context.Context) {

		// Passing a blank state will return all machines
		machines, _ := apiClient.ListMachines(appName, "")
		appHealthyMutex.Lock()
		defer appHealthyMutex.Unlock()

		for _, machine := range machines {

			if machine.State != "started" {
				appHealthy = false
				return
			}
		}
		appHealthy = true
	})
}

func loop(ctx context.Context, p time.Duration, fn func(context.Context)) {
	for ctx.Err() == nil {
		fn(ctx)

		pause.For(ctx, p)
	}
}
