// main.go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/go-redis/redis"

	util "github.com/drax2gma/nodelocker/internal"
)

type JsonResponseType struct {
	Type   string `json:"type"`
	Name   string `json:"name"`
	State  string `json:"state"`
	Owner  string `json:"owner"`
	Expire string `json:"expire"`
}

const (
	C_OK         string = "OK"
	C_ERR        string = "ERROR"
	C_NIL        string = "NIL"
	C_LOCK       string = "locked"
	C_RespHeader string = "application/json"

	C_ERR_JsonConvertData string = "ERR: Error converting LockData to JSON"
	C_ERR_NoNameSpecified string = "ERR: No 'name' parameter specified."
	C_ERR_NoTypeSpecified string = "ERR: No 'type' parameter specified."
	C_ERR_IllegalUser     string = "ERR: Illegal user."
	C_ERR_EnvLockFail     string = "ERR: Env lock unsuccesful."
)

func queryHandler(w http.ResponseWriter, r *http.Request) {

	var jsonRes JsonResponseType

	subjectType := r.URL.Query().Get("type") // 'env' or 'host'
	subjectName := r.URL.Query().Get("name") // name of env or host

	if subjectName == "" {
		// only one parameter allowed, not host and env in the same time
		fmt.Println("-name")
		http.Error(w, C_ERR_NoNameSpecified, http.StatusBadRequest)
		return
	}

	if subjectType == "" {
		// only one parameter allowed, not host and env in the same time
		fmt.Println("-type")
		http.Error(w, C_ERR_NoTypeSpecified, http.StatusBadRequest)
		return
	}

	util.LockDataSUBJECT = fmt.Sprintf("%s:%s", subjectType, subjectName)

	if util.RedisGetLockData() { // got some data on subject

		w.Header().Set("Content-Type", C_RespHeader)

		jsonRes.Type = subjectType
		jsonRes.Name = subjectName
		jsonRes.State = util.LockDataSTATE

		byteData, err := json.Marshal(jsonRes)
		if err != nil {
			http.Error(w, C_ERR_JsonConvertData, http.StatusInternalServerError)
			return
		}

		/* trunk-ignore(golangci-lint/errcheck) */
		w.Write(byteData)

	} else { // no lock data on subject

		w.Header().Set("Content-Type", C_RespHeader)

		jsonRes.Type = subjectType
		jsonRes.Name = subjectName
		jsonRes.State = "unlocked"

		byteData, err := json.Marshal(jsonRes)
		if err != nil {
			http.Error(w, C_ERR_JsonConvertData, http.StatusInternalServerError)
			return
		}

		/* trunk-ignore(golangci-lint/errcheck) */
		w.Write(byteData)
	}
}

func lockHandler(w http.ResponseWriter, r *http.Request) {

	var httpResponse string

	subjectName := r.URL.Query().Get("subject")
	subjectType := r.URL.Query().Get("type")
	lastDay := r.URL.Query().Get("lastday")
	userName := r.URL.Query().Get("user")
	userToken := r.URL.Query().Get("token")

	if subjectName == "" {
		// only one parameter allowed, not host and env in the same time
		fmt.Println("-name")
		http.Error(w, C_ERR_NoNameSpecified, http.StatusBadRequest)
		return
	}

	if subjectType == "" {
		// only one parameter allowed, not host and env in the same time
		fmt.Println("-type")
		http.Error(w, C_ERR_NoTypeSpecified, http.StatusBadRequest)
		return
	}

	if !util.RedisValidUser(userName, userToken) {
		fmt.Println("-user")
		http.Error(w, C_ERR_IllegalUser, http.StatusForbidden)
		return
	}

	if !util.RedisSet(subjectType+":"+subjectName, "state", C_LOCK, util.GetTimeFromNow(lastDay)) {
		http.Error(w, C_ERR_EnvLockFail, http.StatusBadRequest)
		return
	}

	if !util.RedisSet(subjectType+":"+subjectName, "owner", userName, util.GetTimeFromNow(lastDay)) {
		http.Error(w, C_ERR_EnvLockFail, http.StatusBadRequest)
		return
	}

	if !util.RedisSet(subjectType+":"+subjectName, "expire", lastDay, util.GetTimeFromNow(lastDay)) {
		http.Error(w, C_ERR_EnvLockFail, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", C_RespHeader)
	/* trunk-ignore(golangci-lint/errcheck) */
	w.Write([]byte(string(httpResponse)))
}

func unlockHandler(w http.ResponseWriter, r *http.Request) {

	var httpResponse string

	subjectName := r.URL.Query().Get("subject")
	subjectType := r.URL.Query().Get("type")
	userName := r.URL.Query().Get("user")
	userToken := r.URL.Query().Get("token")

	if subjectName == "" {
		// only one parameter allowed, not host and env in the same time
		fmt.Println("+host+env")
		http.Error(w, "ERR: No 'subject' (host or env name) specified.", http.StatusBadRequest)
		return
	}

	if subjectType == "" {
		// only one parameter allowed, not host and env in the same time
		fmt.Println("-host-env")
		http.Error(w, "ERR: No 'type' of subject specified.", http.StatusBadRequest)
		return
	}

	if !util.RedisValidUser(userName, userToken) {
		http.Error(w, "Illegal user.", http.StatusForbidden)
		return
	}

	// UNLOCK function
	util.LockDataSUBJECT = subjectType + ":" + subjectName

	util.RedisDelete()
	httpResponse = fmt.Sprintf("'%s:%s' unlocked successfully by %s.", subjectType, subjectName, userName)

	w.Header().Set("Content-Type", C_RespHeader)
	/* trunk-ignore(golangci-lint/errcheck) */
	w.Write([]byte(string(httpResponse)))

}

func regHandler(w http.ResponseWriter, r *http.Request) {

	userName := r.URL.Query().Get("user")
	userToken := r.URL.Query().Get("token")

	if userName == "" {
		http.Error(w, "Missing 'user' parameter", http.StatusBadRequest)
		return
	}

	if userToken == "" {
		http.Error(w, "Missing 'token' parameter", http.StatusBadRequest)
		return
	}

	// req: user+token
	// check if user exists, if yes, exit
	// new user password into DB = sha1(sha1("USERNAME@nodelocker"))
	// register

	existingUser := util.RedisGet("user", userName)

	if existingUser != nil {
		http.Error(w, "User already created.", http.StatusForbidden)
		return
	}

	if !util.RedisSet("user", userName, userToken, 0) {
		http.Error(w, "ERR: User setup failed.", http.StatusBadRequest)
		return
	} else {
		http.Error(w, "OK: User setup done.", http.StatusAccepted)
		return
	}

}

func adminHandler(w http.ResponseWriter, r *http.Request) {

	var paramResult string

	adminToken := r.URL.Query().Get("token")
	subject := r.URL.Query().Get("subject")
	action := r.URL.Query().Get("action")

	if adminToken == "" {
		paramResult = "Missing 'token' parameter"
	}

	if subject == "user" {
		if action == "purge" {

		} else {
			paramResult = "ERR: Illegal 'action' parameter"
		}

	} else if subject == "env" {
		if action == "create" {

		} else if action == "unlock" {

		} else if action == "terminate" {

		} else {
			paramResult = "ERR: Illegal 'action' parameter"
		}

	} else if subject == "host" {
		if action == "unlock" {

		} else if action == "terminate" {

		} else {
			paramResult = "ERR: Illegal 'action' parameter"
		}

	} else {
		paramResult = "ERR: Illegal 'subject' parameter"
	}

	http.Error(w, paramResult, http.StatusBadRequest)

}

func main() {

	util.RConn = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       1,
	})

	errDb := util.RConn.Ping().Err()
	if errDb == nil {
		fmt.Println("✅ Redis check OK")
	} else {
		fmt.Println("❌ Redis not available, exitting...")
		log.Fatal(errDb.Error())
	}

	r := chi.NewRouter()

	r.Use(middleware.Logger)

	r.Get("/query", queryHandler)
	r.Get("/lock", lockHandler)
	r.Get("/unlock", unlockHandler)
	r.Get("/register", regHandler)
	r.Get("/admin", adminHandler)

	http.Handle("/", r)

	util.ServeTLS(r)
}
