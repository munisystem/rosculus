package postgres

import (
	"database/sql"

	_ "github.com/lib/pq"
)

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

	return db, nil
}

func (p *PostgreSQL) RunQuery(query string) error {
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

	if _, err := tx.Exec(query); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}
