package database

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"time"

	"reports-system/internal/domain/entities"

	_ "github.com/lib/pq"
)

type PostgresDB struct {
	db *sql.DB
}

func NewPostgresDB() (*PostgresDB, error) {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	// Configurar pool de conexões
	maxConns, _ := strconv.Atoi(os.Getenv("DB_MAX_CONNS"))
	if maxConns == 0 {
		maxConns = 10
	}

	idleConns, _ := strconv.Atoi(os.Getenv("DB_IDLE_CONNS"))
	if idleConns == 0 {
		idleConns = 5
	}

	db.SetMaxOpenConns(maxConns)
	db.SetMaxIdleConns(idleConns)
	db.SetConnMaxLifetime(time.Hour)

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &PostgresDB{db: db}, nil
}

func (p *PostgresDB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return p.db.Query(query, args...)
}

func (p *PostgresDB) QueryRow(query string, args ...interface{}) *sql.Row {
	return p.db.QueryRow(query, args...)
}

func (p *PostgresDB) Health() entities.DBHealth {
	stats := p.db.Stats()
	return entities.DBHealth{
		MaxConns:  stats.MaxOpenConnections,
		OpenConns: stats.OpenConnections,
		WaitCount: stats.WaitCount,
		WaitTime:  stats.WaitDuration,
	}
}

func (p *PostgresDB) Close() error {
	return p.db.Close()
}
