package bootstrap

import (
	"auth/internal/infrastructure/app"
	"auth/internal/infrastructure/config"
	"auth/internal/infrastructure/container"
	"auth/internal/infrastructure/logger"
	"auth/internal/infrastructure/middlewares"
	"auth/internal/infrastructure/providers"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

// NewApi initializes the app instance with all necessary configuration.
func NewApi() (*app.App, error) {
	cfg, err := config.New()
	if err != nil {
		return nil, err
	}

	err = logger.New(cfg.Log, cfg.App.Name)

	if err != nil {
		return nil, err
	}

	var ctr *container.Container
	ctr, err = container.New(cfg.Database, cfg.Redis)

	if err != nil {
		return nil, err
	}

	if err = ctr.InitPublisher(cfg.RabbitMQ); err != nil {
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
		func() error {
			if ctr.Publisher != nil {
				return ctr.Publisher.Close()
			}
			return nil
		},
	)

	return appInstance, nil
}

// Run starts the api and manages its lifecycle.
func Run(appInstance *app.App) error {
	srv := newServer(appInstance)

	serverErrors := make(chan error, 1)

	go func() {
		slog.Info("server is starting", "addr", srv.Addr)
		err := srv.ListenAndServe()

		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
		}
	}()

	err := wait(srv, serverErrors)

	if err != nil {
		return err
	}

	appInstance.CloseAll()
	slog.Info("application stopped safely")

	return nil
}

// newServer initializes the HTTP engine and server configuration.
func newServer(appInstance *app.App) *http.Server {
	engine := gin.New()

	middlewares.RegisterMiddlewares(engine, appInstance.Config.App.Env)
	providers.RegisterRoutes(engine, appInstance)

	return &http.Server{
		Addr:    fmt.Sprintf("%s:%s", appInstance.Config.App.Host, appInstance.Config.App.Port),
		Handler: engine,
	}
}

// wait manages the application lifecycle (blocking until signal or error).
func wait(srv *http.Server, serverErrors chan error) error {
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)
	case sig := <-shutdown:
		slog.Info("starting graceful shutdown", "signal", sig.String())

		return shutdownServer(srv)
	}
}

// shutdownServer concern: Specific logic to stop the HTTP server gracefully.
func shutdownServer(srv *http.Server) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		_ = srv.Close()

		return fmt.Errorf("could not stop server gracefully: %w", err)
	}

	return nil
}
