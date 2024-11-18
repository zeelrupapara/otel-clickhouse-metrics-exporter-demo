package clickhousemetrics

import (
	"database/sql"
	"errors"
	"fmt"
	"net/url"

	_ "github.com/ClickHouse/clickhouse-go/v2"
)

var errConfigInvalidEndpoint = errors.New("endpoint must be url format")

type Config struct {
	Endpoint string `mapstructure:"endpoint"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

const defaultEndpoint = "tcp://localhost:9000"
const defaultUsername = "otel"
const defaultPassword = "otel"
const defaultDatabase = "default"

func (cfg *Config) buildDSN() (string, error) {
	dsnURL, err := url.Parse(cfg.Endpoint)
	if err != nil {
		return "", fmt.Errorf("%w: %s", errConfigInvalidEndpoint, err.Error())
	}

	// Use database from config if not specified in path, or if config is not default.
	dsnURL.Path = defaultDatabase

	// Override username and password if specified in config.
	if cfg.Username != "" {
		dsnURL.User = url.UserPassword(cfg.Username, string(cfg.Password))
	}
	return dsnURL.String(), nil
}

var driverName = "clickhouse" // for testing

func (cfg *Config) buildDB() (*sql.DB, error) {
	dsn, err := cfg.buildDSN()
	if err != nil {
		return nil, err
	}

	// ClickHouse sql driver will read clickhouse settings from the DSN string.
	// It also ensures defaults.
	// See https://github.com/ClickHouse/clickhouse-go/blob/08b27884b899f587eb5c509769cd2bdf74a9e2a1/clickhouse_std.go#L189
	conn, err := sql.Open(driverName, dsn)
	if err != nil {
		return nil, err
	}

	return conn, nil
}
