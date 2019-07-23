package main

import (
	"database/sql"
	"time"

	"github.com/palantir/stacktrace"
)

type entry struct {
	ID   int `json:"id"`
	From int `json:"from"`
	To   int `json:"to"`
}

func clockIn(db *sql.DB, email string) (err error) {
	tx, err := db.Begin()
	defer tx.Rollback()
	if err != nil {
		return stacktrace.Propagate(err, "failed to begin a transaction")
	}

	var state string
	err = db.QueryRow("SELECT state FROM user_states WHERE user = ?", email).Scan(&state)
	if err != nil {
		return stacktrace.Propagate(err, "couldn't find a row in user_states for specified user")
	}

	if state == "I" {
		return // already clocked in
	}

	_, err = db.Exec("UPDATE user_states SET state = 'I', since_unix_s = ?1 WHERE user = ?2", time.Now().Unix(), email)
	if err != nil {
		return stacktrace.Propagate(err, "failed to update user state")
	}

	tx.Commit()
	return
}

func clockOut(db *sql.DB, email string) (err error) {
	tx, err := db.Begin()
	defer tx.Rollback()
	if err != nil {
		return stacktrace.Propagate(err, "failed to begin a transaction")
	}

	var state string
	var since int
	err = db.QueryRow("SELECT state, since_unix_s FROM user_states WHERE user = ?", email).Scan(&state, &since)
	if err != nil {
		return stacktrace.Propagate(err, "couldn't find a row in user_states for specified user")
	}

	if state == "O" {
		return // already clocked out
	}

	now := time.Now().Unix() // so that it doesn't change between the next two lines
	_, err = db.Exec("INSERT INTO entries (user, from_unix_s, to_unix_s) VALUES (?1, ?2, ?3)", email, since, now)
	if err != nil {
		return stacktrace.Propagate(err, "failed to insert an entry")
	}
	_, err = db.Exec("UPDATE user_states SET state = 'O', since_unix_s = ?1 WHERE user = ?2", now, email)
	if err != nil {
		return stacktrace.Propagate(err, "failed to update user state")
	}

	tx.Commit()
	return
}

func listEntries(db *sql.DB, email string) (entries []entry, err error) {
	entries = []entry{}
	entry := entry{}
	rows, err := db.Query("SELECT rowid, from_unix_s, to_unix_s FROM entries WHERE user = ?", email)
	if err != nil {
		err = stacktrace.Propagate(err, "failed to list entries")
		return
	}
	for rows.Next() {
		rows.Scan(&entry.ID, &entry.From, &entry.To)
		entries = append(entries, entry)
	}
	return
}
