package main

import (
	"context"
	"log"
	"time"

	"github.com/superfly/flyctl/api"
)

func startAppIfStopped(ctx context.Context, started chan bool) {
	if isHealthy() {
		close(started)
		return
	}

	machines, _ := apiClient.ListMachines(appName, "")

	for _, machine := range machines {
		input := api.StartMachineInput{
			AppID: appName,
			ID:    machine.ID,
		}

		if machine.State == "stopped" || machine.State == "exited" {
			log.Printf("Starting %s", machine.ID)

			apiClient.StartMachine(input)
		}
	}

	loop(ctx, time.Second, func(c context.Context) {
		if isHealthy() {
			close(started)
			return
		}
	})

}

func isHealthy() bool {
	appHealthyMutex.RLock()
	status := appHealthy
	appHealthyMutex.RUnlock()
	return status
}
