package health

import "sync"

var (
	mu      sync.RWMutex
	healthy bool
)

func Is() bool {
	mu.RLock()
	status := healthy
	mu.RUnlock()

	return status
}

func Set() {
	mu.Lock()
	defer mu.Unlock()

	healthy = true
}

func Unset() {
	mu.Lock()
	defer mu.Unlock()

	healthy = false
}
