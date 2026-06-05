package db

import (
	"database/sql"
	"log/slog"

	_ "github.com/mattn/go-sqlite3"
)

// Open открывает соединение с SQLite и применяет схему.
func Open(path string) (*sql.DB, error) {
	database, err := sql.Open("sqlite3", path+"?_foreign_keys=on")
	if err != nil {
		return nil, err
	}
	if err = database.Ping(); err != nil {
		return nil, err
	}
	if err = migrate(database); err != nil {
		return nil, err
	}
	slog.Info("database connected", "path", path)
	return database, nil
}

const schema = `
CREATE TABLE IF NOT EXISTS books (
    id          TEXT PRIMARY KEY,
    title       TEXT NOT NULL,
    author      TEXT NOT NULL,
    isbn        TEXT NOT NULL UNIQUE,
    year        INTEGER NOT NULL,
    status      TEXT NOT NULL DEFAULT 'Available' CHECK(status IN ('Available','Issued'))
);

CREATE TABLE IF NOT EXISTS users (
    id                TEXT PRIMARY KEY,
    name              TEXT NOT NULL,
    email             TEXT NOT NULL UNIQUE,
    registration_date TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS issues (
    id          TEXT PRIMARY KEY,
    book_id     TEXT NOT NULL REFERENCES books(id),
    user_id     TEXT NOT NULL REFERENCES users(id),
    issue_date  TEXT NOT NULL,
    due_date    TEXT NOT NULL,
    return_date TEXT
);
`

func migrate(database *sql.DB) error {
	_, err := database.Exec(schema)
	return err
}
