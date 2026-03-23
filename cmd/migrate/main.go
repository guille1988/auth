package main

import (
	"api/internal/bootstrap"
	"api/internal/infrastructure/logger"
	"flag"
)

func main() {
	fresh := flag.Bool("fresh", false, "drop all tables and migrate again from scratch")
	flag.Parse()

	migration, err := bootstrap.NewMigration()

	if err != nil {
		logger.Fatal(err)
	}

	err = bootstrap.RunMigration(migration, *fresh)

	if err != nil {
		logger.Fatal(err)
	}
}
