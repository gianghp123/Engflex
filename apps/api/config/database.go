package config

import (
	"fmt"
	"net/url"
	"time"

	"engflex-api/internal/utils"
)

// DatabaseConfig holds the PostgreSQL connection settings.
// Values are read from DB_* environment variables (see apps/api/.env).
type DatabaseConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	Name            string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// LoadDatabaseConfig reads the database configuration from the environment.
// Defaults match apps/api/.env for local development.
func LoadDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		Host:            utils.GetEnv("DB_HOST", "localhost"),
		Port:            utils.GetEnv("DB_PORT", "5777"),
		User:            utils.GetEnv("DB_USER", "postgres"),
		Password:        utils.GetEnv("DB_PASSWORD", "postgres"),
		Name:            utils.GetEnv("DB_NAME", "engflex"),
		SSLMode:         utils.GetEnv("DB_SSLMODE", "disable"),
		MaxOpenConns:    utils.GetEnvInt("DB_MAX_OPEN_CONNS", 20),
		MaxIdleConns:    utils.GetEnvInt("DB_MAX_IDLE_CONNS", 10),
		ConnMaxLifetime: time.Duration(utils.GetEnvInt("DB_CONN_MAX_LIFETIME_SEC", 1800)) * time.Second,
		ConnMaxIdleTime: time.Duration(utils.GetEnvInt("DB_CONN_MAX_IDLE_SEC", 300)) * time.Second,
	}
}

// DSN returns a PostgreSQL connection string suitable for pgx / database/sql:
//
//	postgres://user:password@host:port/name?sslmode=disable
func (c DatabaseConfig) DSN() string {
	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(c.User, c.Password),
		Host:   fmt.Sprintf("%s:%s", c.Host, c.Port),
		Path:   "/" + c.Name,
	}
	q := u.Query()
	q.Set("sslmode", c.SSLMode)
	u.RawQuery = q.Encode()
	return u.String()
}

// LogValue returns a DSN with the password redacted, safe for startup logs.
func (c DatabaseConfig) LogValue() string {
	u := &url.URL{
		Scheme: "postgres",
		User:   url.User(c.User),
		Host:   fmt.Sprintf("%s:%s", c.Host, c.Port),
		Path:   "/" + c.Name,
	}
	q := u.Query()
	q.Set("sslmode", c.SSLMode)
	u.RawQuery = q.Encode()
	return u.String()
}
