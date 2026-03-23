package database

import (
	"api/internal/infrastructure/config"
	"database/sql"
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// NewConnection creates a GORM database connection using the specified driver and connection details.
func NewConnection(connection config.DatabaseConnection) (*gorm.DB, error) {
	host := connection.Host
	port := connection.Port
	user := connection.Username
	password := connection.Password
	databaseName := connection.Database
	maxIdleConnections := connection.MaxIdleConnections
	maxOpenConnections := connection.MaxOpenConnections
	driver := connection.Driver

	var dialect gorm.Dialector

	switch driver {
	case config.MySQLDriver:
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", user, password, host, port, databaseName)
		dialect = mysql.Open(dsn)
	case config.PostgresDriver:
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC", host, user, password, databaseName, port)
		dialect = postgres.Open(dsn)
	case config.Sqlite:
		dialect = sqlite.Open(databaseName)
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", driver)
	}

	database, err := gorm.Open(dialect, &gorm.Config{})

	if err != nil {
		return nil, err
	}

	var rawDatabase *sql.DB
	rawDatabase, err = database.DB()

	if err == nil {
		rawDatabase.SetMaxIdleConns(maxIdleConnections)
		rawDatabase.SetMaxOpenConns(maxOpenConnections)
	}

	return database, nil
}
