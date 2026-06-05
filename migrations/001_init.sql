-- migrations/001_init.sql

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
