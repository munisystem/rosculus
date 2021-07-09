package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

const RETRY = 10

type PostgreSQL struct {
	ConnectionURL string
	db            *sql.DB
}

func Initialize(ConnectionURL string) *PostgreSQL {
	return &PostgreSQL{ConnectionURL: ConnectionURL}
}

func (p *PostgreSQL) connection() (*sql.DB, error) {
	if p.db != nil {
		if err := p.db.Ping(); err == nil {
			return p.db, nil
		}
		p.db.Close()
	}

	db, err := sql.Open("postgres", p.ConnectionURL)
	if err != nil {
		return nil, err
	}

	if !waitReady(db) {
		return nil, errors.New("failed to connect PostgreSQL")
	}

	return db, nil
}

func (p *PostgreSQL) RunQueries(queries []string) error {
	db, err := p.connection()

	if err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		tx.Rollback()
	}()

	for _, query := range queries {
		if _, err = tx.Exec(query); err != nil {
			return err
		}
	}

	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

func waitReady(db *sql.DB) bool {
	ready := false
	for i := 0; i < RETRY; i++ {
		log.Println("wait until PostgreSQL is ready...")
		if err := db.Ping(); err == nil {
			ready = true
			break
		}
		time.Sleep(30 * time.Second)
	}
	fmt.Print("\n")

	// If not ready PostgreSQL after 5m, then close connection.
	if !ready {
		db.Close()
	}

	return ready
}
