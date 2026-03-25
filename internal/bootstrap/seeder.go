package bootstrap

import (
	"auth/internal/domain/user"
	"auth/internal/infrastructure/app"
	"auth/internal/infrastructure/config"
	"auth/internal/infrastructure/container"
	"auth/internal/infrastructure/logger"
	"fmt"
	"log/slog"
	"time"
)

// Seeder defines the interface for all seeding operations.
type Seeder interface {
	Run() error
}

// SeederApp holds the dependencies for seeding operations.
type SeederApp struct {
	*app.App
	Seeders []Seeder
}

// NewSeeder initializes the app instance with all necessary configuration for seeding.
func NewSeeder() (*SeederApp, error) {
	cfg, err := config.New()
	if err != nil {
		return nil, err
	}

	err = logger.New(cfg.Log)

	if err != nil {
		return nil, err
	}

	var ctr *container.Container
	ctr, err = container.New(cfg.Database, cfg.Redis)

	if err != nil {
		return nil, err
	}

	appInstance := &app.App{
		Config:    cfg,
		Container: ctr,
	}

	appInstance.AddCloser(
		func() error {
			db, _ := ctr.DefaultConnection.DB()
			return db.Close()
		},
		func() error {
			return ctr.Redis.Close()
		},
	)

	var seeders = []Seeder{
		user.NewSeeder(appInstance),
	}

	return &SeederApp{
		App:     appInstance,
		Seeders: seeders,
	}, nil
}

// RunSeeder executes the seeder logic.
func RunSeeder(seeder *SeederApp) error {
	start := time.Now()

	for _, s := range seeder.Seeders {
		err := s.Run()

		if err != nil {
			return err
		}
	}

	elapsed := time.Since(start).Seconds()
	slog.Info("seeding completed", "duration", fmt.Sprintf("%.2fs", elapsed))

	seeder.CloseAll()

	return nil
}
