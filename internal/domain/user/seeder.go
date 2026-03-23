package user

import (
	"auth/internal/domain/user/model"
	"auth/internal/infrastructure/app"
	"auth/internal/infrastructure/config"
	"log/slog"
)

type Seeder struct {
	app *app.App
}

func NewSeeder(app *app.App) *Seeder {
	return &Seeder{
		app: app,
	}
}

func (seeder *Seeder) Run() error {
	if seeder.app.Config.App.Env == config.ProductionEnv {
		slog.Info("seeding is disabled in production environment")

		return nil
	}

	_, err := model.NewFactory(seeder.app.Container.DefaultConnection).Create(100)

	if err != nil {
		return err
	}

	return nil
}
