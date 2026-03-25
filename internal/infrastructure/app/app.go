package app

import (
	"auth/internal/infrastructure/config"
	"auth/internal/infrastructure/container"
)

type App struct {
	Config    *config.Config
	Container *container.Container
	closers   []func() error
}

// AddCloser registers cleanup functions to be executed when the application closes.
func (app *App) AddCloser(functions ...func() error) {
	app.closers = append(app.closers, functions...)
}

// CloseAll executes all registered cleanup functions in reverse order.
func (app *App) CloseAll() {
	for i := len(app.closers) - 1; i >= 0; i-- {
		_ = app.closers[i]()
	}
}
