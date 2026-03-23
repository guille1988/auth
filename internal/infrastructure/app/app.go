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

// AddCloser registers a cleanup function to be executed when the application closes.
func (app *App) AddCloser(fn func() error) {
	app.closers = append(app.closers, fn)
}

// CloseAll executes all registered cleanup functions in reverse order.
func (app *App) CloseAll() {
	for i := len(app.closers) - 1; i >= 0; i-- {
		_ = app.closers[i]()
	}
}
