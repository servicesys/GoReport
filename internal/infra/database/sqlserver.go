package database

import (
	"database/sql"
	"fmt"
	"os"
	"reports-system/internal/domain/entities"
	"strconv"
	"time"

	_ "github.com/denisenkom/go-mssqldb"
)

type SqlServerDB struct {
	db *sql.DB
}

func (p *SqlServerDB) NewDB() (entities.Database, error) {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	connStr := fmt.Sprintf("sqlserver://%s:%s@%s:%s?database=%s", user, password, host, port, dbname)
	fmt.Println("Connecting to SQL Server with connection string:", connStr)
	db, err := sql.Open("sqlserver", connStr)
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

	return &SqlServerDB{db: db}, nil
}

func (p *SqlServerDB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return p.db.Query(query, args...)
}

func (p *SqlServerDB) QueryRow(query string, args ...interface{}) *sql.Row {
	return p.db.QueryRow(query, args...)
}

func (p *SqlServerDB) Health() entities.DBHealth {
	stats := p.db.Stats()
	return entities.DBHealth{
		MaxConns:  stats.MaxOpenConnections,
		OpenConns: stats.OpenConnections,
		WaitCount: stats.WaitCount,
		WaitTime:  stats.WaitDuration,
	}
}

func (p *SqlServerDB) Close() error {
	return p.db.Close()
}
