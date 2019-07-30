package main

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/palantir/stacktrace"
)

type entry struct {
	ID    int  `json:"id"`
	From  int  `json:"from"`
	To    int  `json:"to"`
	Valid bool `json:"valid"`
}

func disqualify(db *sql.DB) {
	rows, err := db.Query("SELECT user, since_unix_s FROM user_states WHERE state = 'I'")
	if err != nil {
		fmt.Println(stacktrace.Propagate(err, "failed to select users to disqualify"))
		return
	}
	type userSince struct {
		user  string
		since int
	}
	toDisq := []userSince{}
	for rows.Next() {
		var us userSince
		rows.Scan(&us.user, &us.since)
		toDisq = append(toDisq, us)
	}
	for _, x := range toDisq {
		_, err = db.Exec("INSERT INTO entries (user, from_unix_s, to_unix_s, valid) VALUES (?1, ?2, ?3, 0)", x.user, x.since, time.Now().Unix())
		if err != nil {
			fmt.Println(stacktrace.Propagate(err, "failed to add disqualifying entry for "+x.user))
		}
	}
	_, err = db.Exec("UPDATE user_states SET state = 'O', since_unix_s = ? WHERE state = 'I'", time.Now().Unix())
	if err != nil {
		fmt.Println(stacktrace.Propagate(err, "failed to clock out disqualified users"))
	}
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
		return stacktrace.Propagate(err, "ffailed to find a row in user_states for specified user")
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
		return stacktrace.Propagate(err, "failed to find a row in user_states for specified user")
	}

	if state == "O" {
		return // already clocked out
	}

	now := time.Now().Unix() // so that it doesn't change between the next two lines
	_, err = db.Exec("INSERT INTO entries (user, from_unix_s, to_unix_s, valid) VALUES (?1, ?2, ?3, 1)", email, since, now)
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

func editEntry(db *sql.DB, id, from, to int) (err error) {
	_, err = db.Exec("UPDATE entries SET from_unix_s = ?1, to_unix_s = ?2 WHERE rowid = ?3", from, to, id)
	err = stacktrace.Propagate(err, "failed to edit entry")
	return
}

func deleteEntry(db *sql.DB, id int) (err error) {
	_, err = db.Exec("DELETE FROM entries WHERE rowid = ?", id)
	err = stacktrace.Propagate(err, "failed to delete entry")
	return
}

func listEntries(db *sql.DB, email string) (entries []entry, err error) {
	entries = []entry{}
	entry := entry{}
	rows, err := db.Query("SELECT rowid, from_unix_s, to_unix_s, valid FROM entries WHERE user = ?", email)
	if err != nil {
		err = stacktrace.Propagate(err, "failed to list entries")
		return
	}
	for rows.Next() {
		rows.Scan(&entry.ID, &entry.From, &entry.To, &entry.Valid)
		entries = append(entries, entry)
	}
	return
}
