package storage

import (
    "fmt"

    "gorm.io/driver/mysql"
    "gorm.io/driver/postgres"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
)

// Dialect represents a supported database backend.
type Dialect string

const (
    DialectSQLite   Dialect = "sqlite"
    DialectPostgres Dialect = "postgres"
    DialectMySQL    Dialect = "mysql"
)

// Config describes how to connect to a database.
type Config struct {
    Dialect Dialect
    DSN     string
}

// NewDB creates a *gorm.DB using the provided configuration.
func NewDB(cfg Config) (*gorm.DB, error) {
    switch cfg.Dialect {
    case DialectSQLite:
        return gorm.Open(sqlite.Open(cfg.DSN), &gorm.Config{})
    case DialectPostgres:
        return gorm.Open(postgres.Open(cfg.DSN), &gorm.Config{})
    case DialectMySQL:
        return gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{})
    default:
        return nil, fmt.Errorf("unsupported dialect %s", cfg.Dialect)
    }
}
