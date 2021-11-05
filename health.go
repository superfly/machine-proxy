package main

import (
	"context"
	"log"
	"sync"
	"time"
)

var (
	appHealthyMutex sync.RWMutex
	appHealthy      bool
)

func checkAppHealth(ctx context.Context, cfg *config) {
	log.Println("entered checkAppHealth")
	defer log.Println("exited checkAppHealth")

	loop(ctx, time.Second, func(ctx context.Context) {
		// Passing a blank state will return all machines
		machines, err := apiClient.ListMachines(cfg.appName, "")
		if err != nil {
			log.Printf("failed checking app health: %v")

			return
		}

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
