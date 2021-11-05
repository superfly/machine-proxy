package main

import (
	"errors"
	"os"
)

type config struct {
	addr        string
	upstream    string
	appName     string
	accessToken string
}

func configFromEnv() (*config, error) {
	addr, ok := os.LookupEnv("ADDR")
	if !ok || addr == "" {
		addr = ":8080"
	}

	upstream, ok := os.LookupEnv("UPSTREAM")
	if !ok || upstream == "" {
		return nil, errors.New("$UPSTREAM not defined")
	}

	appName, ok := os.LookupEnv("APP_NAME")
	if !ok || appName == "" {
		return nil, errors.New("$APP_NAME not defined")
	}

	accessToken, ok := os.LookupEnv("FLY_ACCESS_TOKEN")
	if !ok || appName == "" {
		return nil, errors.New("$FLY_ACCESS_TOKEN not defined")
	}

	return &config{
		addr:        addr,
		upstream:    upstream,
		appName:     appName,
		accessToken: accessToken,
	}, nil
}
