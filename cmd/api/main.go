package main

import (
	"api/internal/bootstrap"
	"api/internal/infrastructure/logger"
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
