NOW_RFC3339 = $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

all: build

build:
	go build -o bin/machine-proxy -ldflags="-X 'github.com/superfly/machine-proxy/buildinfo.buildDate=$(NOW_RFC3339)'" . 
