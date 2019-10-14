package main

import (
	"crypto/rand"
	"crypto/subtle"
	"database/sql"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/palantir/stacktrace"
	"golang.org/x/crypto/argon2"
)

type Uid int
type Sid string

type onlineUser struct {
	UID   Uid `json:"uid"`
	Since int `json:"since"`
}

func createUser(db *sql.DB, email, password string, admin bool) (uid Uid, err error) {
	// TODO: add email confirmation

	tx, err := db.Begin()
	rollback := func() {
		err = tx.Rollback()
		if err != nil {
			fmt.Println(stacktrace.Propagate(err, "failed to roll back transaction"))
		}
	}
	if err != nil {
		return -1, stacktrace.Propagate(err, "failed to begin a transaction")
	}

	err = db.QueryRow("SELECT 1 FROM users WHERE email = ?", email).Scan()
	if err != sql.ErrNoRows {
		rollback()
		return -1, stacktrace.Propagate(err, "user already exists")
	}

	salt := make([]byte, 8)
	rand.Read(salt)
	hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 1, 16)

	adminInt := 0
	if admin {
		adminInt = 1
	}

	_, err = db.Exec(
		`INSERT INTO users (email, password_hash, password_salt, admin)
		  VALUES (?1, ?2, ?3, ?4)`, email, hash, salt, adminInt)
	if err != nil {
		rollback()
		return -1, stacktrace.Propagate(err, "failed to insert a row into the users table")
	}

	err = db.QueryRow(`SELECT uid FROM users WHERE email = ?`, email).Scan(&uid)
	if err != nil {
		rollback()
		return -1, stacktrace.Propagate(err, "failed to get uid")
	}

	_, err = db.Exec(
		`INSERT INTO user_states (uid, state, since_unix_s)
			VALUES (?1, ?2, ?3)`, uid, "O", time.Now().Unix())
	if err != nil {
		rollback()
		return uid, stacktrace.Propagate(err, "failed to insert a row into the user_states table")
	}

	return uid, stacktrace.Propagate(tx.Commit(), "failed to commit transaction")
}

func checkPassword(db *sql.DB, uid Uid, password string) (ok bool) {
	var savedHash, salt []byte
	err := db.QueryRow("SELECT password_hash, password_salt FROM users WHERE uid = ?", uid).Scan(&savedHash, &salt)
	if err != nil {
		// user doesn't exist
		return false
	}

	computedHash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 1, 16)
	if subtle.ConstantTimeCompare(computedHash, savedHash) == 1 {
		return true
	} else {
		return false
	}
}

func checkSession(db *sql.DB, sid Sid) (ok bool) {
	err := db.QueryRow("SELECT 1 FROM sessions WHERE sid = ?1 AND expires_unix_s >= ?2", sid, time.Now().Unix()).Scan()
	return err != sql.ErrNoRows
}

func getUserBySession(db *sql.DB, sid Sid) (uid Uid, err error) {
	db.QueryRow("SELECT uid FROM sessions WHERE sid = ?1 AND expires_unix_s >= ?2", sid, time.Now().Unix()).Scan(&uid)
	return uid, err
}

func createSession(db *sql.DB, uid Uid, expireAfter time.Duration) (sid Sid, err error) {
	sidRaw := make([]byte, 18)
	rand.Read(sidRaw)
	sid = Sid(base64.StdEncoding.EncodeToString(sidRaw))
	expires := time.Now().Add(expireAfter).Unix()
	_, err = db.Exec(
		`INSERT INTO sessions (sid, uid, expires_unix_s)
			VALUES (?1, ?2, ?3)`, sid, uid, expires)
	return sid, err
}

func cleanSessions(db *sql.DB) (err error) {
	_, err = db.Exec("DELETE FROM sessions WHERE expires_unix_s < ?", time.Now().Unix())
	return err
}

func checkAdmin(db *sql.DB, uid Uid) (admin bool, err error) {
	err = db.QueryRow("SELECT admin FROM users WHERE uid = ?", uid).Scan(&admin)
	return admin, err
}

func countOnlineUsers(db *sql.DB) (onlineUsers int, err error) {
	err = db.QueryRow("SELECT COUNT(*) FROM user_states WHERE state = 'I'").Scan(&onlineUsers)
	return onlineUsers, err
}

func listOnlineUsers(db *sql.DB) (onlineUsers []onlineUser, err error) {
	rows, err := db.Query("SELECT uid, since_unix_s FROM user_states WHERE state = 'I'")
	if err != nil {
		return onlineUsers, stacktrace.Propagate(err, "failed to get online users")
	}

	for rows.Next() {
		var ou onlineUser
		err = rows.Scan(&ou.UID, &ou.Since)
		if err != nil {
			return onlineUsers, stacktrace.Propagate(err, "failed to scan row")
		}
		onlineUsers = append(onlineUsers, ou)
	}

	return onlineUsers, nil
}

func emailToUID(db *sql.DB, email string) (uid Uid, err error) {
	err = db.QueryRow("SELECT uid FROM users WHERE email = ?", email).Scan(&uid)
	return uid, err
}

func uidToEmail(db *sql.DB, uid Uid) (email string, err error) {
	err = db.QueryRow("SELECT email FROM users WHERE uid = ?", uid).Scan(&email)
	return email, err
}
