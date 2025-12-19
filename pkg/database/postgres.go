package database

import (
	"database/sql"
	"fmt"
	"time"

	"go-gin-sqlx-template/config"

	"github.com/XSAM/otelsql"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

type Database struct {
	DB *sqlx.DB
}

func NewPostgresDatabase(cfg config.Config) (*Database, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
	)

	// Open database connection with otelsql instrumentation
	db, err := otelsql.Open("postgres", dsn,
		otelsql.WithAttributesGetter(GetAttrs),
		otelsql.WithAttributes(semconv.DBSystemPostgreSQL),
		otelsql.WithSpanOptions(otelsql.SpanOptions{
			OmitConnResetSession: true,
			OmitRows:             true,
			DisableQuery:         true,
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Verify connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Wrap with sqlx
	sqlxDB := sqlx.NewDb(db, "postgres")

	// Set connection pool settings
	sqlxDB.SetMaxOpenConns(25)
	sqlxDB.SetMaxIdleConns(5)
	sqlxDB.SetConnMaxLifetime(5 * time.Minute)

	return &Database{DB: sqlxDB}, nil
}

func (d *Database) Close() error {
	if d.DB != nil {
		return d.DB.Close()
	}
	return nil
}

func (d *Database) HealthCheck() error {
	return d.DB.Ping()
}

func (d *Database) NewTransactionManager() Transactor {
	return NewTransactionManager(d.DB)
}

func SetMapSqlNamed(args map[string]any) map[string]any {
	m := make(map[string]any, len(args))
	for k, v := range args {
		m[k] = sql.Named(k, v)
	}
	return m
}
