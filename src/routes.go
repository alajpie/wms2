package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/AndrewBurian/powermux"
	"github.com/palantir/stacktrace"
)

type key int

const (
	sessionKey key = iota
	userKey
)

func routes(mux *powermux.ServeMux, env env) {
	mux.Route("/authorize").PostFunc(env.authorize)
	mux.Route("/status").MiddlewareFunc(env.requireSession).GetFunc(env.status)
	mux.Route("/entries").MiddlewareFunc(env.requireSession).GetFunc(env.entries)
	mux.Route("/entries/:id").MiddlewareFunc(env.requireSession).MiddlewareFunc(env.requireAdmin).PutFunc(env.entriesEdit)
	mux.Route("/entries/:id").MiddlewareFunc(env.requireSession).MiddlewareFunc(env.requireAdmin).DeleteFunc(env.entriesDelete)
	mux.Route("/clock/in").MiddlewareFunc(env.requireSession).PutFunc(env.clockIn)
	mux.Route("/clock/out").MiddlewareFunc(env.requireSession).PutFunc(env.clockOut)
}

func do400(w http.ResponseWriter) {
	w.WriteHeader(400)
	w.Write([]byte("400 Bad Request"))
}

func do401(w http.ResponseWriter) {
	w.WriteHeader(401)
	w.Write([]byte("401 Unauthorized"))
}

func do500(w http.ResponseWriter) {
	w.WriteHeader(500)
	w.Write([]byte("500 Internal Server Error"))
}

func (env *env) requireAdmin(w http.ResponseWriter, r *http.Request, n func(http.ResponseWriter, *http.Request)) {
	user, ok := r.Context().Value(userKey).(string)
	if !ok {
		fmt.Println(stacktrace.NewError("malformed context, use requireSession first"))
		do500(w)
		return
	}

	admin, err := checkAdmin(env.db, user)
	if err != nil {
		fmt.Println(stacktrace.Propagate(err, "checkAdmin failed"))
		do500(w)
		return
	}
	if !admin {
		do401(w)
		return
	}

	n(w, r)
}

func (env *env) requireSession(w http.ResponseWriter, r *http.Request, n func(http.ResponseWriter, *http.Request)) {
	h := r.Header.Get("Authorization")
	minLen := len("Bearer ") + 24 // length of session id
	if len(h) < minLen || h[:7] != "Bearer " {
		do401(w)
		return
	}

	id := h[7:]
	user, err := getUserBySession(env.db, id)
	if err != nil {
		do401(w)
		return
	}

	ctx := context.WithValue(r.Context(), sessionKey, id)
	ctx = context.WithValue(ctx, userKey, user)
	n(w, r.WithContext(ctx))
}

func (env *env) status(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(userKey).(string)
	if !ok {
		fmt.Println(stacktrace.NewError("malformed context"))
		do500(w)
		return
	}

	info := struct {
		State string `json:"state"`
		Since int    `json:"since"`
	}{}

	err := env.db.QueryRow("SELECT state, since_unix_s FROM user_states WHERE user = ?", user).Scan(&info.State, &info.Since)
	if err != nil {
		fmt.Println(stacktrace.Propagate(err, "failed to get user info"))
		do500(w)
		return
	}

	js, _ := json.Marshal(info)
	w.Write([]byte(js))
}

func (env *env) clockIn(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(userKey).(string)
	if !ok {
		fmt.Println(stacktrace.NewError("malformed context"))
		do500(w)
		return
	}

	err := clockIn(env.db, user)
	if err != nil {
		fmt.Println(stacktrace.Propagate(err, "failed to clock in"))
		do500(w)
		return
	}
}

func (env *env) clockOut(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(userKey).(string)
	if !ok {
		fmt.Println(stacktrace.NewError("malformed context"))
		do500(w)
		return
	}

	err := clockOut(env.db, user)
	if err != nil {
		fmt.Println(stacktrace.Propagate(err, "failed to clock out"))
		do500(w)
		return
	}
}

func (env *env) entries(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(userKey).(string)
	if !ok {
		fmt.Println(stacktrace.NewError("malformed context"))
		do500(w)
		return
	}
	entries, err := listEntries(env.db, user)
	if err != nil {
		fmt.Println(stacktrace.Propagate(err, ""))
		do500(w)
		return
	}
	js, _ := json.Marshal(entries)
	w.Write([]byte(js))
}

func (env *env) entriesEdit(w http.ResponseWriter, r *http.Request) {
	strID := powermux.PathParam(r, "id")
	id, err := strconv.Atoi(strID)
	if err != nil {
		do400(w)
		return
	}
	r.ParseForm()
	strFrom := r.Form.Get("from")
	from, err := strconv.Atoi(strFrom)
	if err != nil {
		do400(w)
		return
	}
	strTo := r.Form.Get("to")
	to, err := strconv.Atoi(strTo)
	if err != nil {
		do400(w)
		return
	}

	if to < from {
		do400(w)
		return
	}

	err = editEntry(env.db, id, from, to)
	if err != nil {
		fmt.Println(stacktrace.Propagate(err, ""))
		do500(w)
		return
	}
}

func (env *env) entriesDelete(w http.ResponseWriter, r *http.Request) {
	strID := powermux.PathParam(r, "id")
	id, err := strconv.Atoi(strID)
	if err != nil {
		do400(w)
		return
	}
	err = deleteEntry(env.db, id)
	if err != nil {
		fmt.Println(stacktrace.Propagate(err, ""))
		do500(w)
		return
	}
}

func (env *env) authorize(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	email := r.Form.Get("email")
	password := r.Form.Get("password")
	ok := checkPassword(env.db, email, password)
	if !ok {
		do401(w)
		return
	}

	id, err := createSession(env.db, email, time.Hour*24*31)
	if err != nil {
		fmt.Println(stacktrace.Propagate(err, "createSession failed"))
		do500(w)
		return
	}

	w.Write([]byte(id))
}
