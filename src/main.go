package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/AndrewBurian/powermux"
	"github.com/palantir/stacktrace"
)

type env struct {
	db *sql.DB
}

func disqualifier(db *sql.DB) {
	for {
		now := time.Now()
		cutoff := time.Date(now.Year(), now.Month(), now.Day(), 20, 0, 0, 0, now.Location())
		time.Sleep(time.Until(cutoff))
		disqualify(db)
	}
}

func main() {
	const init = `
	CREATE TABLE users (
		uid INTEGER PRIMARY KEY AUTOINCREMENT, -- so that they don't repeat
		email TEXT,
		password_hash BLOB,
		password_salt BLOB,
		admin INTEGER CHECK(admin IN (0, 1)),
		UNIQUE(email)
	);

	CREATE TABLE user_states (
		uid INTEGER,
		state TEXT CHECK(state IN ('I', 'O')),
		since_unix_s, -- see entries.from_unix_s
		FOREIGN KEY (uid) REFERENCES users(uid),
		UNIQUE(uid)
	);

	CREATE TABLE entries (
		eid INTEGER PRIMARY KEY AUTOINCREMENT, -- so that they don't repeat
		uid INTEGER,
		from_unix_s INTEGER, -- "_s" stands for seconds, unlike the JS millisecond unix time
		to_unix_s INTEGER, -- see above, can be null, signifies disqualifed entry
		valid INTEGER CHECK(valid IN (0, 1)),
		FOREIGN KEY (uid) REFERENCES users(uid),
		CHECK(from_unix_s <= to_unix_s)
	);

	CREATE TABLE sessions (
		sid TEXT,
		uid INTEGER,
		expires_unix_s INTEGER, -- see entries.from_unix_s
		FOREIGN KEY (uid) REFERENCES users(uid)
	);

	CREATE INDEX sessions_id ON sessions (sid);
	`
	var db *sql.DB
	if _, err := os.Stat("./wms2.db"); os.IsNotExist(err) {
		// the database hasn't been created yet
		// so we create it...
		db, err = sql.Open("sqlite3", "./wms2.db?mode=rwc")
		if err != nil {
			fmt.Println(stacktrace.Propagate(err, "failed to open the database"))
			return
		}
		defer db.Close()

		// ...and execute the code
		_, err = db.Exec(init)
		if err != nil {
			fmt.Println(stacktrace.Propagate(err, "failed to execute init SQL"))
			return
		}
	} else {
		// the database exists so we assume it's initialised
		db, err = sql.Open("sqlite3", "./wms2.db?mode=rw")
		if err != nil {
			fmt.Println(stacktrace.Propagate(err, "failed to open the database"))
			return
		}
		defer db.Close()
	}

	db.Exec(`PRAGMA foreign_keys = on;`)

	cleanSessions(db)
	createUser(db, "test@invalid", "hunter2", false)
	createUser(db, "admin@invalid", "hunter2", true)

	go disqualifier(db)

	mux := powermux.NewServeMux()
	env := env{db}
	routes(mux, env)
	err := http.ListenAndServe(":3000", mux)
	fmt.Println(stacktrace.Propagate(err, ""))
}
