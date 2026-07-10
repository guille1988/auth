package bootstrap

import (
	"auth/internal/infrastructure/app"
	"auth/internal/infrastructure/config"
	"auth/internal/infrastructure/container"
	"auth/internal/infrastructure/logger"
	"auth/internal/infrastructure/middlewares"
	"auth/internal/infrastructure/providers"
	"auth/internal/infrastructure/providers/messaging"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/guille1988/go-app-shared/messaging/kafka/constants"
	"github.com/guille1988/go-app-shared/messaging/kafka/dtos"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
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

	ctr.Publisher, err = setupPublisher(cfg.Kafka)

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
		func() error {
			if ctr.Publisher != nil {
				flushCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()

				err = ctr.Publisher.Flush(flushCtx)

				if err != nil {
					slog.Error("kafka flush on shutdown failed", "error", err)
				}

				return ctr.Publisher.Close()
			}

			return nil
		},
	)

	return appInstance, nil
}

func setupPublisher(cfg config.KafkaConfig) (messaging.Publisher, error) {
	publisher := messaging.NewKafkaPublisher(cfg.Brokers)

	if err := publisher.Register(dtos.WelcomeEmail{}, messaging.Route{
		RoutingKey: constants.RouteUserCreated,
	}); err != nil {
		_ = publisher.Close()
		return nil, err
	}

	if err := publisher.Register(dtos.UserLoggedIn{}, messaging.Route{
		RoutingKey: constants.RouteUserLoggedIn,
	}); err != nil {
		_ = publisher.Close()
		return nil, err
	}

	if err := publisher.Register(dtos.StressEmail{}, messaging.Route{
		RoutingKey: constants.RouteStressTest,
	}); err != nil {
		_ = publisher.Close()
		return nil, err
	}

	return publisher, nil
}

// Run starts the api (HTTP and gRPC listeners) and manages its lifecycle.
func Run(appInstance *app.App) error {
	srv := newServer(appInstance)

	serverErrors := make(chan error, 2)

	go func() {
		slog.Info("server is starting", "addr", srv.Addr)
		err := srv.ListenAndServe()

		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
		}
	}()

	grpcServer, err := startGRPCServer(appInstance, serverErrors)

	if err != nil {
		_ = srv.Close()
		return err
	}

	err = wait(srv, grpcServer, serverErrors)

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
func wait(srv *http.Server, grpcServer *grpc.Server, serverErrors chan error) error {
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		/*
		 One listener failed: stop the other before returning, so CloseAll
		 never runs with a live listener.
		*/
		_ = srv.Close()
		grpcServer.Stop()

		return fmt.Errorf("server error: %w", err)
	case sig := <-shutdown:
		slog.Info("starting graceful shutdown", "signal", sig.String())

		return shutdownServers(srv, grpcServer)
	}
}

/*
shutdownServers stops both listeners gracefully, in parallel: k8s gives the
pod terminationGracePeriodSeconds (60s), and the two 30s budgets would exceed
it if run sequentially.
*/
func shutdownServers(srv *http.Server, grpcServer *grpc.Server) error {
	var waitGroup sync.WaitGroup
	var httpErr error

	waitGroup.Add(2)

	go func() {
		defer waitGroup.Done()
		httpErr = shutdownHTTPServer(srv)
	}()

	go func() {
		defer waitGroup.Done()
		shutdownGRPCServer(grpcServer)
	}()

	waitGroup.Wait()

	return httpErr
}

// shutdownHTTPServer concern: Specific logic to stop the HTTP server gracefully.
func shutdownHTTPServer(srv *http.Server) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		_ = srv.Close()

		return fmt.Errorf("could not stop server gracefully: %w", err)
	}

	return nil
}
