package entities

import (
	"database/sql"
	"time"
)

type DBHealth struct {
	MaxConns  int           `json:"max_conns"`
	OpenConns int           `json:"open_conns"`
	WaitCount int64         `json:"wait_count"`
	WaitTime  time.Duration `json:"wait_time"`
}

type Database interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Health() DBHealth
	Close() error
}
