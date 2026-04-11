package bootstrap

import (
	"auth/internal/infrastructure/app"
	"auth/internal/infrastructure/config"
	"auth/internal/infrastructure/container"
	"auth/internal/infrastructure/logger"
	"auth/internal/infrastructure/middlewares"
	"auth/internal/infrastructure/providers"
	"net/http"

	"github.com/gin-gonic/gin"
)

// NewTestingApi initializes the app optimized for tests.
func NewTestingApi(cfg *config.Config) (*app.App, error) {
	cfg.App.Env = config.TestingEnv

	err := logger.New(cfg.Log, cfg.App.Name)

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

	appInstance.AddCloser(func() error {
		db, _ := ctr.DefaultConnection.DB()

		if db != nil {
			_ = db.Close()
		}

		if ctr.Redis != nil {
			_ = ctr.Redis.Close()
		}

		if ctr.Publisher != nil {
			_ = ctr.Publisher.Close()
		}

		return nil
	})

	return appInstance, nil
}

// NewTestingHandler returns the Gin engine without an HTTP server.
func NewTestingHandler(appInstance *app.App) http.Handler {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	middlewares.RegisterMiddlewares(engine, appInstance.Config.App.Env)
	providers.RegisterRoutes(engine, appInstance)

	return engine
}
