package clickhouse

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

type ClickHouseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
}

var ClickHouseClient driver.Conn

func InitClickHouseNoDB(cfg ClickHouseConfig) {
	var err error
	var conn driver.Conn

	for i := 0; i < 10; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		slog.Info("Connecting to ClickHouse", "host", cfg.Host, "port", cfg.Port)

		opts := &clickhouse.Options{
			Addr: []string{fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)},
			Auth: clickhouse.Auth{
				Database: "",
				Username: cfg.User,
				Password: cfg.Password,
			},
			Settings: clickhouse.Settings{
				"max_execution_time":                        60,
				"output_format_native_write_json_as_string": 1,
			},
			DialTimeout:      5 * time.Second,
			MaxOpenConns:     10,
			MaxIdleConns:     5,
			ConnMaxLifetime:  time.Hour,
			ConnOpenStrategy: clickhouse.ConnOpenInOrder,
		}

		if cfg.Port == "9440" {
			opts.TLS = &tls.Config{
				InsecureSkipVerify: false,
			}
		}

		conn, err = clickhouse.Open(opts)

		if err == nil {
			err = conn.Ping(ctx)
		}

		if err == nil {
			slog.Info("ClickHouse connected successfully")
			ClickHouseClient = conn
			return
		}

		slog.Warn("ClickHouse connection failed", "attempt", i+1, "error", err)
		time.Sleep(3 * time.Second)
	}

	slog.Error("Could not connect to ClickHouse after multiple attempts", "error", err)
	os.Exit(1)
}

func InitClickHouse(cfg ClickHouseConfig) {
	var err error
	var conn driver.Conn

	for i := 0; i < 10; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		slog.Info("Connecting to ClickHouse", "host", cfg.Host, "port", cfg.Port)

		opts := &clickhouse.Options{
			Addr: []string{fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)},
			Auth: clickhouse.Auth{
				Database: cfg.Database,
				Username: cfg.User,
				Password: cfg.Password,
			},
			Settings: clickhouse.Settings{
				"max_execution_time":                        60,
				"output_format_native_write_json_as_string": 1,
			},
			DialTimeout:      5 * time.Second,
			MaxOpenConns:     10,
			MaxIdleConns:     5,
			ConnMaxLifetime:  time.Hour,
			ConnOpenStrategy: clickhouse.ConnOpenInOrder,
		}

		if cfg.Port == "9440" {
			opts.TLS = &tls.Config{
				InsecureSkipVerify: false,
			}
		}

		conn, err = clickhouse.Open(opts)

		if err == nil {
			err = conn.Ping(ctx)
		}

		if err == nil {
			slog.Info("ClickHouse connected successfully")
			ClickHouseClient = conn
			return
		}

		slog.Warn("ClickHouse connection failed", "attempt", i+1, "error", err)
		time.Sleep(3 * time.Second)
	}

	slog.Error("Could not connect to ClickHouse after multiple attempts", "error", err)
	os.Exit(1)
}

func CreateEventsTable(ctx context.Context) error {
	query := `
	CREATE TABLE IF NOT EXISTS events (
		id String,
		event_type String,
		aggregate_type String,
		aggregate_id String,
		organization_id String,
		user_id String,
		user_name String,
		request_id String,
		payload String,
		metadata String,
		timestamp DateTime64(3, 'UTC'),
		processed_by Array(String),
		created_at DateTime64(3, 'UTC') DEFAULT now64(3, 'UTC')
	) ENGINE = MergeTree()
	PARTITION BY toYYYYMM(timestamp)
	ORDER BY (organization_id, aggregate_id, timestamp)
	SETTINGS index_granularity = 8192
	`

	if err := ClickHouseClient.Exec(ctx, query); err != nil {
		return fmt.Errorf("failed to create events table: %w", err)
	}

	slog.Info("Events table created or already exists in ClickHouse")
	return nil
}

func CreateDatabase(ctx context.Context, database string) error {
	query := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", database)

	if err := ClickHouseClient.Exec(ctx, query); err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}

	slog.Info("Database created or already exists in ClickHouse", "database", database)
	return nil
}
