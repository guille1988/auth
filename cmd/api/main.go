package main

import (
	"auth/internal/bootstrap"
	"auth/internal/infrastructure/logger"
)

func main() {
	api, err := bootstrap.NewApi()

	if err != nil {
		logger.Fatal(err)
	}

	err = bootstrap.Run(api)

	if err != nil {
		logger.Fatal(err)
	}
}
