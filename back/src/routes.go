package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/AndrewBurian/powermux"
	"github.com/palantir/stacktrace"
)

const apiVersion = -1

type key int

const (
	sidKey key = iota
	uidKey
)

func routes(mux *powermux.ServeMux, env env) {
	mux.Route("/").MiddlewareFunc(env.corsMiddleware)
	mux.Route("/version").GetFunc(env.version)
	mux.Route("/authorize").PostFunc(env.authorize)
	u := mux.Route("/u").MiddlewareFunc(env.requireSession)
	u.Route("/status").GetFunc(env.status)
	u.Route("/entries").GetFunc(env.entries)
	u.Route("/clock/in").PutFunc(env.clockIn)
	u.Route("/clock/out").PutFunc(env.clockOut)
	u.Route("/users/online/count").GetFunc(env.usersOnlineCount)
	a := mux.Route("/a").MiddlewareFunc(env.requireSession).MiddlewareFunc(env.requireAdmin)
	a.Route("/entries/:id").PutFunc(env.entriesEdit)
	a.Route("/entries/:id").DeleteFunc(env.entriesDelete)
	a.Route("/users/:id")
	a.Route("/users/online/list").GetFunc(env.usersOnlineList)
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

func (env *env) version(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(strconv.Itoa(apiVersion)))
}

func (env *env) corsMiddleware(w http.ResponseWriter, r *http.Request, n func(http.ResponseWriter, *http.Request)) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
	if r.Method == "OPTIONS" {
		w.WriteHeader(200)
	} else {
		n(w, r)
	}
}

func (env *env) requireAdmin(w http.ResponseWriter, r *http.Request, n func(http.ResponseWriter, *http.Request)) {
	uid, ok := r.Context().Value(uidKey).(uid_t)
	if !ok {
		fmt.Println(stacktrace.NewError("malformed context, use requireSession first"))
		do500(w)
		return
	}

	admin, err := checkAdmin(env.db, uid)
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

	sid := sid_t(h[7:])
	uid, err := getUserBySession(env.db, sid)
	if err != nil {
		do401(w)
		return
	}

	ctx := context.WithValue(r.Context(), sidKey, sid)
	ctx = context.WithValue(ctx, uidKey, uid)
	n(w, r.WithContext(ctx))
}

func (env *env) status(w http.ResponseWriter, r *http.Request) {
	uid, ok := r.Context().Value(uidKey).(uid_t)
	if !ok {
		fmt.Println(stacktrace.NewError("malformed context"))
		do500(w)
		return
	}

	info := struct {
		State         string `json:"state"`
		Since         int    `json:"since"`
		Online        int    `json:"online"`
		DeltaForMonth int    `json:"deltaForMonth"`
		DeltaForDay   int    `json:"deltaForDay"`
	}{}

	online, err := countOnlineUsers(env.db)
	info.Online = online
	if err != nil {
		fmt.Println(stacktrace.Propagate(err, "failed to count online users"))
		do500(w)
		return
	}

	deltaForMonth, err := getDeltaForMonth(env.db, uid, time.Now())
	info.DeltaForMonth = deltaForMonth
	if err != nil {
		fmt.Println(stacktrace.Propagate(err, "failed to get monthly delta"))
		do500(w)
		return
	}

	deltaForDay, err := getDeltaForDay(env.db, uid, time.Now())
	info.DeltaForDay = deltaForDay
	if err != nil {
		fmt.Println(stacktrace.Propagate(err, "failed to get daily delta"))
		do500(w)
		return
	}

	err = env.db.QueryRow("SELECT state, since_unix_s FROM user_states WHERE uid = ?", uid).Scan(&info.State, &info.Since)
	if err != nil {
		fmt.Println(stacktrace.Propagate(err, "failed to get user info"))
		do500(w)
		return
	}

	js, _ := json.Marshal(info)
	w.Write([]byte(js))
}

func (env *env) clockIn(w http.ResponseWriter, r *http.Request) {
	uid, ok := r.Context().Value(uidKey).(uid_t)
	if !ok {
		fmt.Println(stacktrace.NewError("malformed context"))
		do500(w)
		return
	}

	err := clockIn(env.db, uid)
	if err != nil {
		fmt.Println(stacktrace.Propagate(err, "failed to clock in"))
		do500(w)
		return
	}
}

func (env *env) clockOut(w http.ResponseWriter, r *http.Request) {
	uid, ok := r.Context().Value(uidKey).(uid_t)
	if !ok {
		fmt.Println(stacktrace.NewError("malformed context"))
		do500(w)
		return
	}

	err := clockOut(env.db, uid)
	if err != nil {
		fmt.Println(stacktrace.Propagate(err, "failed to clock out"))
		do500(w)
		return
	}
}

func (env *env) entries(w http.ResponseWriter, r *http.Request) {
	uid, ok := r.Context().Value(uidKey).(uid_t)
	if !ok {
		fmt.Println(stacktrace.NewError("malformed context"))
		do500(w)
		return
	}

	entries, err := listEntries(env.db, uid)
	if err != nil {
		fmt.Println(stacktrace.Propagate(err, ""))
		do500(w)
		return
	}

	js, _ := json.Marshal(entries)
	w.Write([]byte(js))
}

func (env *env) entriesEdit(w http.ResponseWriter, r *http.Request) {
	strEID := powermux.PathParam(r, "eid")
	intEID, err := strconv.Atoi(strEID)
	eid := eid_t(intEID)
	if err != nil {
		do400(w)
		return
	}

	err = r.ParseForm()
	if err != nil {
		do400(w)
		return
	}
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

	err = editEntry(env.db, eid, from, to)
	if err != nil {
		fmt.Println(stacktrace.Propagate(err, ""))
		do500(w)
		return
	}
}

func (env *env) entriesDelete(w http.ResponseWriter, r *http.Request) {
	strEID := powermux.PathParam(r, "eid")
	intEID, err := strconv.Atoi(strEID)
	eid := eid_t(intEID)
	if err != nil {
		do400(w)
		return
	}

	err = deleteEntry(env.db, eid)
	if err != nil {
		fmt.Println(stacktrace.Propagate(err, ""))
		do500(w)
		return
	}
}

func (env *env) usersOnlineCount(w http.ResponseWriter, r *http.Request) {
	onlineUsers, err := countOnlineUsers(env.db)
	if err != nil {
		fmt.Print(stacktrace.Propagate(err, "failed to count online users"))
		do500(w)
		return
	}

	w.Write([]byte(strconv.Itoa(onlineUsers)))
}

func (env *env) usersOnlineList(w http.ResponseWriter, r *http.Request) {
	onlineUsers, err := listOnlineUsers(env.db)
	if err != nil {
		fmt.Print(stacktrace.Propagate(err, "failed to list online users"))
		do500(w)
		return
	}

	if onlineUsers == nil { // empty list
		onlineUsers = make([]onlineUser, 0) // doesn't marshal to null
	}

	js, _ := json.Marshal(onlineUsers)
	w.Write([]byte(js))
}

func (env *env) authorize(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		do500(w)
		return
	}

	type form struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	f := form{}
	json.Unmarshal(body, &f)

	uid, err := emailToUID(env.db, f.Email)
	if err != nil {
		do401(w)
		return
	}

	ok := checkPassword(env.db, uid, f.Password)
	if !ok {
		do401(w)
		return
	}

	sid, err := createSession(env.db, uid, time.Hour*24*31)
	if err != nil {
		fmt.Println(stacktrace.Propagate(err, "failed to create a session"))
		do500(w)
		return
	}

	js, _ := json.Marshal(struct {
		Token sid_t `json:"token"`
	}{sid})
	w.Write([]byte(js))
}
