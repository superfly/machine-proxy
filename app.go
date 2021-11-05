package main

import (
	"context"
	"log"
	"time"

	"github.com/superfly/flyctl/api"

	"github.com/superfly/machine-proxy/internal/health"
)

func ensureAppIsRunning(ctx context.Context, cfg *config) {
	if health.Is() {
		return
	}

	machines, err := cfg.client.ListMachines(cfg.appName, "")
	if err != nil {
		return
	}

	for _, machine := range machines {
		input := api.StartMachineInput{
			AppID: cfg.appName,
			ID:    machine.ID,
		}

		if machine.State == "stopped" || machine.State == "exited" {
			log.Printf("starting %s", machine.ID)

			if _, err := cfg.client.StartMachine(input); err != nil {
				log.Printf("failed starting: %v", err)

				return
			}
		}
	}
}

func isAppRunning(ctx context.Context) (bool, error) {
	for ctx.Err() == nil {
		if health.Is() {
			return true, nil
		}

		time.Sleep(100 * time.Millisecond)
	}

	return false, ctx.Err()
}
