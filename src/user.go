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

type onlineUser struct {
	Email string `json:"email"`
	Since int    `json:"since"`
}

func createUser(db *sql.DB, email, password string, admin bool) (err error) {
	// TODO: add email confirmation

	tx, err := db.Begin()
	rollback := func() {
		err = tx.Rollback()
		if err != nil {
			fmt.Println(stacktrace.Propagate(err, "failed to roll back transaction"))
		}
	}
	if err != nil {
		return stacktrace.Propagate(err, "failed to begin a transaction")
	}

	err = db.QueryRow("SELECT 1 FROM users WHERE email = ?", email).Scan()
	if err != sql.ErrNoRows {
		rollback()
		return stacktrace.Propagate(err, "user already exists")
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
		return stacktrace.Propagate(err, "failed to insert a row into the users table")
	}

	_, err = db.Exec(
		`INSERT INTO user_states (user, state, since_unix_s)
			VALUES (?1, ?2, ?3)`, email, "O", time.Now().Unix())
	if err != nil {
		rollback()
		return stacktrace.Propagate(err, "failed to insert a row into the user_states table")
	}

	return stacktrace.Propagate(tx.Commit(), "failed to commit transaction")
}

func checkPassword(db *sql.DB, email, password string) (ok bool) {
	var savedHash, salt []byte
	err := db.QueryRow("SELECT password_hash, password_salt FROM users WHERE email = ?", email).Scan(&savedHash, &salt)
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

func checkSession(db *sql.DB, id string) (ok bool) {
	err := db.QueryRow("SELECT 1 FROM sessions WHERE id = ?1 AND expires_unix_s >= ?2", id, time.Now().Unix()).Scan()
	return err != sql.ErrNoRows
}

func getUserBySession(db *sql.DB, id string) (email string, err error) {
	db.QueryRow("SELECT user FROM sessions WHERE id = ?1 AND expires_unix_s >= ?2", id, time.Now().Unix()).Scan(&email)
	return email, err
}

func createSession(db *sql.DB, email string, expireAfter time.Duration) (id string, err error) {
	idRaw := make([]byte, 18)
	rand.Read(idRaw)
	id = base64.StdEncoding.EncodeToString(idRaw)
	expires := time.Now().Add(expireAfter).Unix()
	_, err = db.Exec(
		`INSERT INTO sessions (id, user, expires_unix_s)
			VALUES (?1, ?2, ?3)`, id, email, expires)
	return id, err
}

func cleanSessions(db *sql.DB) (err error) {
	_, err = db.Exec("DELETE FROM sessions WHERE expires_unix_s < ?", time.Now().Unix())
	return err
}

func checkAdmin(db *sql.DB, email string) (admin bool, err error) {
	err = db.QueryRow("SELECT admin FROM users WHERE email = ?", email).Scan(&admin)
	return admin, err
}

func countOnlineUsers(db *sql.DB) (onlineUsers int, err error) {
	err = db.QueryRow("SELECT COUNT(user) FROM user_states WHERE state = 'I'").Scan(&onlineUsers)
	return onlineUsers, err
}

func listOnlineUsers(db *sql.DB) (onlineUsers []onlineUser, err error) {
	rows, err := db.Query("SELECT user, since_unix_s FROM user_states WHERE state = 'I'")
	if err != nil {
		return onlineUsers, stacktrace.Propagate(err, "failed to get online users")
	}

	for rows.Next() {
		var ou onlineUser
		err = rows.Scan(&ou.Email, &ou.Since)
		if err != nil {
			return onlineUsers, stacktrace.Propagate(err, "failed to scan row")
		}
		onlineUsers = append(onlineUsers, ou)
	}

	return onlineUsers, nil
}
