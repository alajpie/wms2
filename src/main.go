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
		email TEXT PRIMARY KEY,
		password_hash BLOB,
		password_salt BLOB,
		admin INTEGER CHECK(admin IN (0, 1))
	);

	CREATE TABLE user_states (
		user TEXT,
		state TEXT CHECK(state IN ('I', 'O')),
		since_unix_s, -- see entries.from_unix_s
		FOREIGN KEY (user) REFERENCES users(email)
	);

	CREATE TABLE entries (
		rowid INTEGER PRIMARY KEY AUTOINCREMENT, -- so that rowids don't repeat
		user TEXT,
		from_unix_s INTEGER, -- "_s" stands for seconds, unlike the JS millisecond unix time
		to_unix_s INTEGER, -- see above, can be null, signifies disqulifed entry
		valid INTEGER CHECK(valid IN (0, 1)),
		FOREIGN KEY (user) REFERENCES users(email),
		CHECK(from_unix_s <= to_unix_s)
	);

	CREATE TABLE sessions (
		id TEXT,
		user TEXT,
		expires_unix_s INTEGER, -- see entries.from_unix_s
		FOREIGN KEY (user) REFERENCES users(email)
	);

	CREATE INDEX sessions_id ON sessions (id);
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
