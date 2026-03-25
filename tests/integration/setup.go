package integration

import (
	"auth/internal/bootstrap"
	"auth/internal/infrastructure/app"
	"auth/internal/infrastructure/config"
	"auth/internal/infrastructure/database"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/docker/go-connections/nat"
	"github.com/golang-migrate/migrate/v4"
	"github.com/testcontainers/testcontainers-go/modules/mysql"
	"github.com/testcontainers/testcontainers-go/modules/rabbitmq"
	"github.com/testcontainers/testcontainers-go/modules/redis"
)

var TestHandler http.Handler
var TestConfig *config.Config

// RunTests handles the integration tests setup, execution, and cleanup.
func RunTests(test *testing.M) {
	ctx := context.Background()

	TestConfig = setupConfig()
	mysqlInstance := setupDatabaseContainer(ctx, TestConfig)
	redisInstance := setupRedisContainer(ctx, TestConfig)
	rabbitInstance := setupRabbitContainer(ctx, TestConfig)

	setupApplication(TestConfig)

	code := test.Run()

	_ = mysqlInstance.Terminate(ctx)
	_ = redisInstance.Terminate(ctx)
	_ = rabbitInstance.Terminate(ctx)
	os.Exit(code)
}

// setupConfig initializes and returns the application configuration.
func setupConfig() *config.Config {
	cfg, _ := config.New()
	return cfg
}

// setupDatabaseContainer starts a MySQL container and updates the configuration with dynamic connection details.
func setupDatabaseContainer(ctx context.Context, cfg *config.Config) *mysql.MySQLContainer {
	databaseConfig := cfg.Database.Connections[config.Default]
	testDatabaseName := databaseConfig.Database + "_test"

	mysqlInstance, err := mysql.Run(ctx, "mysql:8.0",
		mysql.WithDatabase(testDatabaseName),
		mysql.WithUsername(databaseConfig.Username),
		mysql.WithPassword(databaseConfig.Password),
	)

	if err != nil {
		panic(err)
	}

	if mysqlInstance == nil {
		panic("mysqlInstance is nil, you should not see this")
	}

	host, _ := mysqlInstance.Host(ctx)
	port, _ := mysqlInstance.MappedPort(ctx, "3306")

	databaseConfig.Host = host
	databaseConfig.Port = port.Port()
	databaseConfig.Database = testDatabaseName
	cfg.Database.Connections[config.Default] = databaseConfig

	return mysqlInstance
}

// setupRedisContainer starts a Redis container and updates the configuration with dynamic connection details.
func setupRedisContainer(ctx context.Context, cfg *config.Config) *redis.RedisContainer {
	redisContainer, err := redis.Run(ctx, "redis:7-alpine")

	if err != nil {
		panic(err)
	}

	if redisContainer == nil {
		panic("redisContainer is nil, you should not see this")
	}

	host, _ := redisContainer.Host(ctx)
	port, _ := redisContainer.MappedPort(ctx, nat.Port(cfg.Redis.Port))

	cfg.Redis.Host = host
	cfg.Redis.Port = port.Port()

	return redisContainer
}

// setupRabbitContainer starts a RabbitMQ container and updates the configuration with dynamic connection details.
func setupRabbitContainer(ctx context.Context, cfg *config.Config) *rabbitmq.RabbitMQContainer {
	rabbitContainer, err := rabbitmq.Run(ctx, "rabbitmq:3-management-alpine",
		rabbitmq.WithAdminPassword(cfg.RabbitMQ.Password),
		rabbitmq.WithAdminUsername(cfg.RabbitMQ.User),
	)

	if err != nil {
		panic(err)
	}

	if rabbitContainer == nil {
		panic("rabbitContainer is nil, you should not see this")
	}

	host, _ := rabbitContainer.Host(ctx)
	port, _ := rabbitContainer.MappedPort(ctx, nat.Port(cfg.RabbitMQ.Port))

	cfg.RabbitMQ.Host = host
	cfg.RabbitMQ.Port = port.Port()

	return rabbitContainer
}

// setupApplication initializes the API instance, runs database migrations, and sets the global test handler.
func setupApplication(cfg *config.Config) {
	appInstance, err := bootstrap.NewTestingApi(cfg)

	if err != nil {
		panic(err)
	}

	var migration *migrate.Migrate
	migration, err = database.NewMigration(*cfg, config.Default)

	if err != nil {
		panic(err)
	}

	err = database.Migrate(migration)

	if err != nil {
		panic(err)
	}

	TestHandler = bootstrap.NewTestingHandler(appInstance)
}

// RefreshDatabase resets the database to a clean state by running down and then up migrations.
func RefreshDatabase() {
	migration, err := database.NewMigration(*TestConfig, config.Default)

	if err != nil {
		panic(err)
	}

	err = migration.Down()

	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		panic(err)
	}

	err = database.Migrate(migration)

	if err != nil {
		panic(err)
	}
}

// TestCase is a wrapper that runs RefreshDatabase before executing the actual test logic.
func TestCase(test *testing.T, name string, testFunction func(test *testing.T)) {
	test.Run(name, func(test *testing.T) {
		RefreshDatabase()
		testFunction(test)
	})
}

// GetToken registers a user and returns their access token.
func GetToken() string {
	payload := map[string]string{
		"name":     "Test User",
		"email":    "test@example.com",
		"password": "password123",
	}
	body, _ := json.Marshal(payload)

	request, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")

	response := ExecuteRequest(request)

	var data map[string]any
	_ = json.Unmarshal(response.Body.Bytes(), &data)

	return data["access_token"].(string)
}

func GetApp() (*app.App, error) {
	return bootstrap.NewTestingApi(TestConfig)
}

// ExecuteRequest performs an HTTP request against the global test handler and returns the response recorder.
func ExecuteRequest(request *http.Request) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	TestHandler.ServeHTTP(recorder, request)

	return recorder
}
