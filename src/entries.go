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
	rollback := func() {
		err = tx.Rollback()
		if err != nil {
			fmt.Println(stacktrace.Propagate(err, "failed to roll back transaction"))
		}
	}
	if err != nil {
		return stacktrace.Propagate(err, "failed to begin transaction")
	}

	var state string
	err = db.QueryRow("SELECT state FROM user_states WHERE user = ?", email).Scan(&state)
	if err != nil {
		rollback()
		return stacktrace.Propagate(err, "failed to find a row in user_states for specified user")
	}

	if state == "I" {
		rollback()
		return nil // already clocked in
	}

	_, err = db.Exec("UPDATE user_states SET state = 'I', since_unix_s = ?1 WHERE user = ?2", time.Now().Unix(), email)
	if err != nil {
		rollback()
		return stacktrace.Propagate(err, "failed to update user state")
	}

	return stacktrace.Propagate(tx.Commit(), "failed to commit transaction")
}

func clockOut(db *sql.DB, email string) (err error) {
	tx, err := db.Begin()
	rollback := func() {
		err = tx.Rollback()
		if err != nil {
			fmt.Println(stacktrace.Propagate(err, "failed to roll back transaction"))
		}
	}
	if err != nil {
		return stacktrace.Propagate(err, "failed to begin transaction")
	}

	var state string
	var since int
	err = db.QueryRow("SELECT state, since_unix_s FROM user_states WHERE user = ?", email).Scan(&state, &since)
	if err != nil {
		rollback()
		return stacktrace.Propagate(err, "failed to find a row in user_states for specified user")
	}

	if state == "O" {
		rollback()
		return nil // already clocked out
	}

	now := time.Now().Unix() // so that it doesn't change between the next two SQL statements
	_, err = db.Exec("INSERT INTO entries (user, from_unix_s, to_unix_s, valid) VALUES (?1, ?2, ?3, 1)", email, since, now)
	if err != nil {
		rollback()
		return stacktrace.Propagate(err, "failed to insert an entry")
	}
	_, err = db.Exec("UPDATE user_states SET state = 'O', since_unix_s = ?1 WHERE user = ?2", now, email)
	if err != nil {
		rollback()
		return stacktrace.Propagate(err, "failed to update user state")
	}

	return stacktrace.Propagate(tx.Commit(), "failed to commit transaction")
}

func editEntry(db *sql.DB, id, from, to int) (err error) {
	_, err = db.Exec("UPDATE entries SET from_unix_s = ?1, to_unix_s = ?2 WHERE rowid = ?3", from, to, id)
	return stacktrace.Propagate(err, "failed to edit entry")
}

func deleteEntry(db *sql.DB, id int) (err error) {
	_, err = db.Exec("DELETE FROM entries WHERE rowid = ?", id)
	return stacktrace.Propagate(err, "failed to delete entry")
}

func listEntries(db *sql.DB, email string) (days map[int64][]entry, err error) {
	ens := []entry{}
	en := entry{}
	rows, err := db.Query("SELECT rowid, from_unix_s, to_unix_s, valid FROM entries WHERE user = ?", email)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to list entries")
	}
	for rows.Next() {
		err = rows.Scan(&en.ID, &en.From, &en.To, &en.Valid)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to scan row")
		}
		ens = append(ens, en)
	}
	days = make(map[int64][]entry)
	for _, x := range ens {
		date := time.Unix(int64(x.From), 0)
		key := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location()).Unix()
		days[key] = append(days[key], x)
	}

	return days, nil
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
		return delta, stacktrace.Propagate(err, "failed to get entries in date range")
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
	return delta, nil
}
