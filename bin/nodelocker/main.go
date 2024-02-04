// main.go
package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/go-redis/redis"

	util "github.com/drax2gma/internal"
)

type LockRecord struct {
	id     string // name of specific host or env
	state  string // 'locked', 'maintenance' or 'terminated' (latter by admin only)
	owner  string // Who locked it? Username
	expire string // Last locking date in YYYYMMDD format
}

const (
	OK   string = "OK"
	ERR  string = "ERROR"
	NIL  string = "NIL"
	LOCK string = "locked"
)

var (
	httpResponse string
	retval       string
	status       string
)

func envHandler(w http.ResponseWriter, r *http.Request) {

	action := r.URL.Query().Get("action")
	envName := r.URL.Query().Get("env")
	lastDay := r.URL.Query().Get("lastday")
	userName := r.URL.Query().Get("user")
	userToken := r.URL.Query().Get("token")

	// req: env (as filter)
	if action == "query" {
		retval, status = util.RedisGet("env:" + envName)
		if status == NIL || status == ERR {
			httpResponse = NIL
		} else {
			httpResponse = status
		}

		// req: env+lastday+user+token
	} else if action == "lock" {
		if !util.RedisValidUser(userName, userToken) {
			retval = ERR
			status = "Illegal user"
		}
		retval, status = util.RedisSet("env:"+envName, "state", LOCK)
		if status != OK {
			http.Error(w, "Env lock unsuccesful.", http.StatusBadRequest)
			return
		}

		retval, status = util.RedisSet("env:"+envName, "owner", userName)
		if status != OK {
			http.Error(w, "Env lock unsuccesful.", http.StatusBadRequest)
			return
		}

		retval, status = util.RedisSet("env:"+envName, "expire", util.GetTimeFromNow(lastDay))
		if status != OK {
			http.Error(w, "Env lock unsuccesful.", http.StatusBadRequest)
			return
		}

		// req: env+user+token
	} else if action == "unlock" {
		if !util.RedisValidUser(userName, userToken) {
			status = ERR
			retval = "Illegal user"
		}
		// UNLOCK function
		retval, status = util.RedisSet("env:"+envName, "", "", time.Now().Add(time.Second))
		if status != OK {
			http.Error(w, "Env unlock failed.", http.StatusBadRequest)
			return
		}

	} else {
		res := fmt.Sprintf("Illegal action on environment %s by %s\n", envName, userName)
		util.RedisLog(res)
	}

	w.Header().Set("Content-Type", "application/text")

	w.Write([]byte(string(httpResponse)))
}

func hostHandler(w http.ResponseWriter, r *http.Request) {

	action := r.URL.Query().Get("action")
	hostName := r.URL.Query().Get("host")
	lastDay := r.URL.Query().Get("lastday")
	userName := r.URL.Query().Get("user")
	userToken := r.URL.Query().Get("token")

	// req: host (as filter)
	if action == "query" {
		retval, status = util.RedisGet("host:" + hostName)
		if status == NIL || status == ERR {
			httpResponse = NIL
		} else {
			httpResponse = status
		}

		// req: env+lastday+user+token
	} else if action == "lock" {
		if !util.RedisValidUser(userName, userToken) {
			retval = ERR
			status = "Illegal user"
		}
		retval, status = util.RedisSet("host:"+hostName, "state", LOCK)
		if status != OK {
			http.Error(w, "Env lock unsuccesful.", http.StatusBadRequest)
			return
		}

		retval, status = util.RedisSet("host:"+hostName, "owner", userName)
		if status != OK {
			http.Error(w, "Env lock unsuccesful.", http.StatusBadRequest)
			return
		}

		retval, status = util.RedisSet("host:"+hostName, "expire", util.GetTimeFromNow(lastDay))
		if status != OK {
			http.Error(w, "Env lock unsuccesful.", http.StatusBadRequest)
			return
		}

		// req: host+user+token
	} else if action == "unlock" {
		if !util.RedisValidUser(userName, userToken) {
			status = ERR
			retval = "Illegal user"
		}
		// UNLOCK function
		retval, status = util.RedisSet("host:"+hostName, "", "", time.Now().Add(time.Second))
		if status != OK {
			http.Error(w, "Env unlock failed.", http.StatusBadRequest)
			return
		}

	} else {
		res := fmt.Sprintf("Illegal action on host %s by %s\n", hostName, userName)
		util.RedisLog(res)
	}

	w.Header().Set("Content-Type", "application/text")

	w.Write([]byte(string(httpResponse)))
}

func userHandler(w http.ResponseWriter, r *http.Request) {

	userName := r.URL.Query().Get("user")
	userToken := r.URL.Query().Get("token")
	action := r.URL.Query().Get("action")

	if userName == "" {
		http.Error(w, "Missing 'user' parameter", http.StatusBadRequest)
		return
	}

	// req: user+token
	if action == "register" {
		// check if user exists, if yes, exit
		// new user password into DB = sha1(sha1("USERNAME@nodelocker"))
		// register
	} else {
		http.Error(w, "Invalid 'action' parameter", http.StatusBadRequest)
		return
	}

	w.Write([]byte("Login successful"))
}

func adminHandler(w http.ResponseWriter, r *http.Request) {

	adminToken := r.URL.Query().Get("token")
	action := r.URL.Query().Get("action")

	if adminToken == "" {
		http.Error(w, "Missing 'token' parameter", http.StatusBadRequest)
		return
	}

	if action == "user-purge" {

	} else if action == "env-create" {

	} else if action == "env-unlock" {

	} else if action == "host-unlock" {

	} else if action == "env-terminate" {

	} else if action == "host-terminate" {

	} else {

	}

	w.Write([]byte("Login successful"))
}

func main() {

	util.RConn = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	errDb := util.rConn.Ping().Err()
	if errDb == nil {
		fmt.Println("✅ Redis check OK")
	} else {
		fmt.Println("❌ Redis not available, exitting...")
		log.Fatal(errDb.Error())
	}

	r := chi.NewRouter()

	r.Use(middleware.Logger)

	r.Get("/env", envHandler)
	r.Get("/host", hostHandler)
	r.Get("/user", userHandler)
	r.Get("/admin", adminHandler)

	http.Handle("/", r)

	/* trunk-ignore(golangci-lint/errcheck) */
	http.ListenAndServe(":3000", nil)
}
