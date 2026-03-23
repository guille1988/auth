package database

import (
	"auth/internal/infrastructure/config"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

const (
	localPath      string = "file:///auth/internal/infrastructure/database/migrations/local"
	testingPath    string = "file:///auth/internal/infrastructure/database/migrations/testing"
	stagingPath    string = "file:///auth/internal/infrastructure/database/migrations/staging"
	productionPath string = "file:///auth/internal/infrastructure/database/migrations/production"
)

// NewMigration returns a migration instance according to the connection name.
func NewMigration(cfg config.Config, connectionName config.ConnectionName) (*migrate.Migrate, error) {
	dbConfig := cfg.Database.Connections[connectionName]
	var dsn string

	switch dbConfig.Driver {
	case config.MySQLDriver:
		dsn = fmt.Sprintf("mysql://%s:%s@tcp(%s:%s)/%s",
			dbConfig.Username, dbConfig.Password, dbConfig.Host, dbConfig.Port, dbConfig.Database)
	case config.PostgresDriver:
		dsn = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
			dbConfig.Username, dbConfig.Password, dbConfig.Host, dbConfig.Port, dbConfig.Database)
	case config.Sqlite:
		dsn = fmt.Sprintf("sqlite3://%s", dbConfig.Database)
	default:
		text := fmt.Sprintf("unsupported database driver: %s", dbConfig.Driver)
		return nil, errors.New(text)
	}

	var path string

	switch cfg.App.Env {
	case config.LocalEnv:
		path = localPath
	case config.TestingEnv:
		path = testingPath
	case config.ProductionEnv:
		path = productionPath
	case config.StagingEnv:
		path = stagingPath
	default:
		text := fmt.Sprintf("unsupported environment for migration: %s", cfg.App.Env)
		return nil, errors.New(text)
	}

	return migrate.New(path, dsn)
}

// Migrate executes the database migration.
func Migrate(migration *migrate.Migrate) error {
	err := migration.Up()

	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}
