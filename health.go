package main

import (
	"context"
	"log"
	"time"

	"github.com/superfly/machine-proxy/internal/health"
)

func checkAppHealth(ctx context.Context, cfg *config) {
	log.Println("entered checkAppHealth")
	defer log.Println("exited checkAppHealth")

	loop(ctx, time.Second, func(ctx context.Context) {
		// Passing a blank state will return all machines
		machines, err := cfg.client.ListMachines(cfg.appName, "")
		if err != nil {
			log.Printf("failed checking app health: %v", err)

			return
		}

		for _, machine := range machines {
			if machine.State != "started" {
				health.Unset()

				return
			}
		}

		health.Set()
	})
}
