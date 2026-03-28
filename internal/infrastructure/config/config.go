package config

import (
	"auth/internal/infrastructure/env"
	"time"

	"github.com/joho/godotenv"
)

// Config represents the application configuration.
type Config struct {
	App      AppConfig
	Database DatabaseConfig
	Log      LogConfig
	Redis    RedisConfig
	Auth     AuthConfig
	RabbitMQ RabbitMQConfig
}

type RabbitMQConfig struct {
	Host     string
	Port     string
	User     string
	Password string
}

// AppConfig represents the application configuration.
type AppConfig struct {
	Name string
	Env  Env
	Host string
	Port string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	Database int
}

type AuthConfig struct {
	JWTSecret               string
	AccessTokenExpire       time.Duration
	RefreshTokenExpire      time.Duration
	EmailVerificationExpire time.Duration
}

type ConnectionName string

const (
	Default ConnectionName = "default"
)

// DatabaseConfig represents the database configuration.
type DatabaseConfig struct {
	Connections map[ConnectionName]DatabaseConnection
}

// DatabaseConnection represents the database connection.
type DatabaseConnection struct {
	Driver             Driver
	Host               string
	Port               string
	Database           string
	Username           string
	Password           string
	MaxIdleConnections int
	MaxOpenConnections int
}

type LogConfig struct {
	Driver LogDriver
	Path   string
	Level  LogLevel
}

type Driver string

const (
	MySQLDriver    Driver = "mysql"
	PostgresDriver Driver = "postgres"
	Sqlite         Driver = "sqlite"
)

type Env string

const (
	LocalEnv      Env = "local"
	TestingEnv    Env = "testing"
	StagingEnv    Env = "staging"
	ProductionEnv Env = "production"
)

type LogLevel string

const (
	DebugLevel LogLevel = "debug"
	InfoLevel  LogLevel = "info"
	WarnLevel  LogLevel = "warn"
	ErrorLevel LogLevel = "error"
)

type LogDriver string

const (
	StdoutFormat LogDriver = "stdout"
	File         LogDriver = "file"
)

// New creates a new configuration instance.
func New() (*Config, error) {
	_ = godotenv.Load()

	config := Config{
		App: AppConfig{
			Name: env.GetEnvAsString("APP_NAME", "auth"),
			Env:  Env(env.GetEnvAsString("APP_ENV", string(LocalEnv))),
			Host: env.GetEnvAsString("APP_HOST", "localhost"),
			Port: env.GetEnvAsString("APP_PORT", "8080"),
		},
		Database: DatabaseConfig{
			Connections: map[ConnectionName]DatabaseConnection{
				Default: {
					Driver:             Driver(env.GetEnvAsString("DB_DRIVER", string(MySQLDriver))),
					Host:               env.GetEnvAsString("DB_HOST", "mysql_auth"),
					Port:               env.GetEnvAsString("DB_PORT", "3306"),
					Database:           env.GetEnvAsString("DB_DATABASE", "auth"),
					Username:           env.GetEnvAsString("DB_USERNAME", "auth"),
					Password:           env.GetEnvAsString("DB_PASSWORD", "auth"),
					MaxIdleConnections: env.GetEnvAsInt("DB_MAX_IDLE_CONNECTIONS", 10),
					MaxOpenConnections: env.GetEnvAsInt("DB_MAX_OPEN_CONNECTIONS", 10),
				},
			},
		},
		Log: LogConfig{
			Driver: LogDriver(env.GetEnvAsString("LOG_DRIVER", string(StdoutFormat))),
			Path:   env.GetEnvAsString("LOG_PATH", "logs/auth.log"),
			Level:  LogLevel(env.GetEnvAsString("LOG_LEVEL", string(InfoLevel))),
		},
		Redis: RedisConfig{
			Host:     env.GetEnvAsString("REDIS_HOST", "redis"),
			Port:     env.GetEnvAsString("REDIS_PORT", "6379"),
			Password: env.GetEnvAsString("REDIS_PASSWORD", "auth"),
			Database: env.GetEnvAsInt("REDIS_DATABASE", 0),
		},
		Auth: AuthConfig{
			JWTSecret:               env.GetEnvAsString("AUTH_JWT_SECRET", "secret"),
			AccessTokenExpire:       time.Duration(env.GetEnvAsInt("AUTH_ACCESS_TOKEN_EXPIRE", 15)) * time.Minute,
			RefreshTokenExpire:      time.Duration(env.GetEnvAsInt("AUTH_REFRESH_TOKEN_EXPIRE", 10080)) * time.Minute,
			EmailVerificationExpire: time.Duration(env.GetEnvAsInt("AUTH_EMAIL_VERIFICATION_EXPIRE", 60)) * time.Minute,
		},
		RabbitMQ: RabbitMQConfig{
			Host:     env.GetEnvAsString("RABBITMQ_HOST", "rabbitmq"),
			Port:     env.GetEnvAsString("RABBITMQ_PORT", "5672"),
			User:     env.GetEnvAsString("RABBITMQ_USER", "guest"),
			Password: env.GetEnvAsString("RABBITMQ_PASSWORD", "guest"),
		},
	}

	return &config, nil
}
