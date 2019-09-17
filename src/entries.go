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
		err = rows.Scan(&us.user, &us.since)
		if err != nil {
			fmt.Print(stacktrace.Propagate(err, "failed to scan row"))
		}
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
	defer func() {
		err = tx.Rollback()
		if err != nil {
			fmt.Println(stacktrace.Propagate(err, "failed to roll back transaction"))
		}
	}()
	if err != nil {
		return stacktrace.Propagate(err, "failed to begin transaction")
	}

	var state string
	err = db.QueryRow("SELECT state FROM user_states WHERE user = ?", email).Scan(&state)
	if err != nil {
		return stacktrace.Propagate(err, "failed to find a row in user_states for specified user")
	}

	if state == "I" {
		return // already clocked in
	}

	_, err = db.Exec("UPDATE user_states SET state = 'I', since_unix_s = ?1 WHERE user = ?2", time.Now().Unix(), email)
	if err != nil {
		return stacktrace.Propagate(err, "failed to update user state")
	}

	err = stacktrace.Propagate(tx.Commit(), "failed to commit transaction")
	return
}

func clockOut(db *sql.DB, email string) (err error) {
	tx, err := db.Begin()
	defer func() {
		err = tx.Rollback()
		if err != nil {
			fmt.Println(stacktrace.Propagate(err, "failed to roll back transaction"))
		}
	}()
	if err != nil {
		return stacktrace.Propagate(err, "failed to begin transaction")
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

	err = stacktrace.Propagate(tx.Commit(), "failed to commit transaction")
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
		err = rows.Scan(&entry.ID, &entry.From, &entry.To, &entry.Valid)
		if err != nil {
			err = stacktrace.Propagate(err, "failed to scan row")
			return
		}
		entries = append(entries, entry)
	}
	return
}

func getDeltaForMonth(db *sql.DB, email string, date time.Time) (delta int, err error) {
	// TODO: account for holidays
	som := time.Date(date.Year(), date.Month(), 0, 0, 0, 0, 0, date.Location())
	eom := time.Date(date.Year(), date.Month()+1, 0, 0, 0, 0, 0, date.Location())
	rows, err := db.Query(
		`SELECT from_unix_s, to_unix_s FROM entries
			WHERE user = ?1 AND valid = 1
			AND from_unix_s > ?2 AND to_unix_s < ?3`, email, som.Unix(), eom.Unix())
	if err != nil {
		err = stacktrace.Propagate(err, "failed to get entries in date range")
		return
	}
	for rows.Next() {
		var from, to int
		rows.Scan(&from, &to)
		delta += to - from
	}
	x := som
	for x.Before(eom) {
		if x.Weekday() != time.Saturday && x.Weekday() != time.Sunday {
			delta -= 8 * 60 * 60
		}
		x = x.Add(time.Hour * 24)
	}
	return
}
