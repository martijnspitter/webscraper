package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	db *sql.DB
}

type Property struct {
	ID        int64
	Address   string
	Source    string // e.g., "rebo", "vesteda", etc.
	FirstSeen time.Time
	LastSeen  time.Time
	Active    bool
}

func New(dbPath string) (*Database, error) {
	// Create the directory if it doesn't exist
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	// Create tables if they don't exist
	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS properties (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            address TEXT NOT NULL,
            source TEXT NOT NULL,
            first_seen DATETIME NOT NULL,
            last_seen DATETIME NOT NULL,
            active BOOLEAN NOT NULL,
            UNIQUE(address, source)
        );
    `)
	if err != nil {
		return nil, err
	}

	return &Database{db: db}, nil
}

func (d *Database) Close() error {
	return d.db.Close()
}

// Add a new property or update existing one
func (d *Database) UpsertProperty(address, source string) error {
	now := time.Now()
	_, err := d.db.Exec(`
        INSERT INTO properties (address, source, first_seen, last_seen, active)
        VALUES (?, ?, ?, ?, TRUE)
        ON CONFLICT(address, source)
        DO UPDATE SET last_seen = ?, active = TRUE
    `, address, source, now, now, now)
	return err
}

// Get all active properties for a source
func (d *Database) GetActiveProperties(source string) ([]Property, error) {
	rows, err := d.db.Query(`
        SELECT id, address, source, first_seen, last_seen, active
        FROM properties
        WHERE source = ? AND active = TRUE
    `, source)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var properties []Property
	for rows.Next() {
		var p Property
		err := rows.Scan(&p.ID, &p.Address, &p.Source, &p.FirstSeen, &p.LastSeen, &p.Active)
		if err != nil {
			return nil, err
		}
		properties = append(properties, p)
	}
	return properties, nil
}

// Mark properties as inactive if they're no longer listed
func (d *Database) MarkInactive(source string, activeAddresses []string) error {
	if len(activeAddresses) == 0 {
		return nil
	}
	query := `
        UPDATE properties
        SET active = FALSE
        WHERE source = ?
        AND active = TRUE
        AND address NOT IN (?` + strings.Repeat(",?", len(activeAddresses)-1) + ")"

	args := make([]interface{}, len(activeAddresses)+1)
	args[0] = source
	for i, addr := range activeAddresses {
		args[i+1] = addr
	}

	_, err := d.db.Exec(query, args...)
	return err
}
