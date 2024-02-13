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

	x "github.com/drax2gma/nodelocker/internal"
)

const (
	C_OK         string = "OK"
	C_ERR        string = "ERROR"
	C_NIL        string = "NIL"
	C_LOCK       string = "locked"
	C_RespHeader string = "application/json"
	C_Secret     string = "XXXXXXX"

	ERR_JsonConvertData      string = "ERR: Error converting LockData to JSON."
	ERR_NoNameSpecified      string = "ERR: No 'name' parameter specified."
	ERR_NoTypeSpecified      string = "ERR: No 'type' parameter specified."
	ERR_NoUserSpecified      string = "ERR: No 'user' parameter specified."
	ERR_NoTokenSpecified     string = "ERR: No 'token' parameter specified."
	ERR_WrongTypeSpecified   string = "ERR: Wrong 'type' specified, must be 'env' or 'host'."
	ERR_IllegalUser          string = "ERR: Illegal user."
	ERR_UserExists           string = "ERR: User already exists."
	ERR_UserSetupFailed      string = "ERR: User setup failed."
	ERR_EnvLockFail          string = "ERR: Env lock unsuccesful."
	ERR_InvalidDateSpecified string = "ERR: Invalid 'lastday' specified, format is: YYYYMMDD."
	ERR_NoAdminPresent       string = "ERR: No 'admin' user present, cannot continue."

	C_TLS bool = true // serve TLS with self-signed cert?
)

func queryHandler(w http.ResponseWriter, r *http.Request) {

	res := x.NewWebResponse()
	c := x.NewCacheData()

	c.Type = r.URL.Query().Get("type") // type of entity, 'env' or 'host'
	c.Name = r.URL.Query().Get("name") // name of entity

	// no 'type' defined in GET request
	if c.Type == "" {

		c.HttpErr = http.StatusBadRequest
		res.Messages = append(res.Messages, ERR_NoTypeSpecified)
	}

	// bad 'type' defined in GET request
	if c.Type != "env" && c.Type != "host" {

		c.HttpErr = http.StatusBadRequest
		res.Messages = append(res.Messages, ERR_WrongTypeSpecified)
	}

	// no 'name' defined in GET request
	if c.Name == "" {

		c.HttpErr = http.StatusBadRequest
		res.Messages = append(res.Messages, ERR_NoNameSpecified)
	}

	if c.HttpErr == 0 { // GET params were good

		if x.RGetLockData(&c) {
			// got some data on entity from lock database
			res.State = c.State
			res.User = c.User
		} else {
			// no lock data on entity
			res.State = "unlocked"
			res.User = ""
		}

		res.Success = true
		res.Messages = []string{"valid_response"}
		res.Type = c.Type
		res.Name = c.Name

	} else { // GET params were problametic a bit

		w.Header().Set("Content-Type", C_RespHeader)
		w.WriteHeader(c.HttpErr)
	}

	byteData, err := json.MarshalIndent(res, "", "    ")
	if err != nil {
		http.Error(w, ERR_JsonConvertData, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", C_RespHeader)
	/* trunk-ignore(golangci-lint/errcheck) */
	w.Write(byteData)
}

func lockHandler(w http.ResponseWriter, r *http.Request) {

	res := x.NewWebResponse()
	c := x.NewCacheData()

	c.Type = r.URL.Query().Get("type")
	c.Name = r.URL.Query().Get("name")
	c.LastDay = r.URL.Query().Get("lastday")
	c.User = r.URL.Query().Get("user")
	c.Token = r.URL.Query().Get("token")

	// Check if init sequence has been made when starting anything as normal user
	if c.User != "admin" {
		if x.RCheckUser(x.C_USER_Exists, "admin", "") == x.C_USER_NotExists {

			c.HttpErr = http.StatusLocked
			res.Messages = append(res.Messages, ERR_NoAdminPresent)
		}
	}

	// no 'type' defined in GET request
	if c.Type == "" {

		c.HttpErr = http.StatusBadRequest
		res.Messages = append(res.Messages, ERR_NoTypeSpecified)
	}

	// bad 'type' defined in GET request
	if !x.IsValidEntityType(c.Type) {

		c.HttpErr = http.StatusBadRequest
		res.Messages = append(res.Messages, ERR_WrongTypeSpecified)
	}

	// no 'name' defined in GET request
	if c.Name == "" {

		c.HttpErr = http.StatusBadRequest
		res.Messages = append(res.Messages, ERR_NoNameSpecified)
	}

	// Is given LASTDAY is a valid date?
	if !x.IsValidDate(c.LastDay) {

		c.HttpErr = http.StatusBadRequest
		res.Messages = append(res.Messages, ERR_InvalidDateSpecified)
	}

	// Is given user valid against DB user? Pwd checking too.
	if x.RCheckUser(x.C_USER_Valid, c.User, c.Token) == x.C_USER_Invalid {

		c.HttpErr = http.StatusForbidden
		res.Messages = append(res.Messages, ERR_IllegalUser)
	}

	//
	if !x.RSetSingle(x.C_UseCacheData, "state", C_LOCK, x.GetTimeFromNow(c.LastDay)) {

		c.HttpErr = http.StatusBadRequest
		res.Messages = append(res.Messages, ERR_EnvLockFail)
	}

	//
	if !x.RSetSingle(x.C_UseCacheData, "user", c.User, x.GetTimeFromNow(c.LastDay)) {

		c.HttpErr = http.StatusBadRequest
		res.Messages = append(res.Messages, ERR_EnvLockFail)
	}

	//
	if !x.RSetSingle(x.C_UseCacheData, "lastday", c.LastDay, x.GetTimeFromNow(c.LastDay)) {

		c.HttpErr = http.StatusBadRequest
		res.Messages = append(res.Messages, ERR_EnvLockFail)
	}

	if c.HttpErr == 0 { // GET params were good

		if x.RGetLockData(&c) {
			// got some data on entity from lock database
			res.State = c.State
			res.User = c.User
		} else {
			// no lock data on entity
			res.State = "unlocked"
			res.User = ""
		}

		res.Success = true
		res.Messages = []string{"valid_response"}
		res.Type = c.Type
		res.Name = c.Name

	} else { // GET params were problametic a bit

		w.Header().Set("Content-Type", C_RespHeader)
		w.WriteHeader(c.HttpErr)
	}

	returnWebResponse(w, c.HttpErr, &res)
}

func unlockHandler(w http.ResponseWriter, r *http.Request) {

	res := x.NewWebResponse()
	c := x.NewCacheData()

	c.Type = r.URL.Query().Get("type")
	c.Name = r.URL.Query().Get("name")
	c.User = r.URL.Query().Get("user")
	c.Token = r.URL.Query().Get("token")

	// Check if init sequence has been made when starting anything as normal user
	if c.User != x.C_ADMIN {
		if x.RCheckUser(x.C_USER_Exists, x.C_ADMIN, "") == x.C_USER_NotExists {

			c.HttpErr = http.StatusLocked
			res.Messages = append(res.Messages, ERR_NoAdminPresent)
		}
	}

	// Check for missing entity type
	if c.Type == "" {
		// only one parameter allowed, not host and env in the same time
		c.HttpErr = http.StatusBadRequest
		res.Messages = append(res.Messages, ERR_NoTypeSpecified)
	}

	// Check for missing entity name
	if c.Name == "" {

		c.HttpErr = http.StatusBadRequest
		res.Messages = append(res.Messages, ERR_NoNameSpecified)
	}

	if x.RCheckUser(x.C_USER_Valid, c.User, c.Token) == x.C_USER_Invalid {

		c.HttpErr = http.StatusForbidden
		res.Messages = append(res.Messages, ERR_IllegalUser)
	}

	if !x.REntityDelete(c.Type, c.Name) {

		c.HttpErr = http.StatusForbidden
		res.Messages = append(res.Messages, ERR_IllegalUser)

	} else {

		c.HttpErr = http.StatusOK
		res.Messages = append(res.Messages, "")
	}

	returnWebResponse(w, c.HttpErr, &res)
}

func regHandler(w http.ResponseWriter, r *http.Request) {

	res := x.NewWebResponse()
	c := x.NewCacheData()

	c.User = r.URL.Query().Get("user")
	c.Token = r.URL.Query().Get("token")

	// Check if init sequence has been made when starting anything as normal user
	if c.User != x.C_ADMIN {
		if x.RCheckUser(x.C_USER_Exists, x.C_ADMIN, "") == x.C_USER_NotExists {

			c.HttpErr = http.StatusLocked
			res.Messages = append(res.Messages, ERR_NoAdminPresent)
		}
	}

	// no 'user' defined in GET request
	if c.User == "" {

		c.HttpErr = http.StatusBadRequest
		res.Messages = append(res.Messages, ERR_NoUserSpecified)
	}

	// no 'token' defined in GET request
	if c.Token == "" {

		c.HttpErr = http.StatusBadRequest
		res.Messages = append(res.Messages, ERR_NoTokenSpecified)
	}

	if x.RCheckUser(x.C_USER_Exists, c.User, "") == x.C_USER_Exists {

		c.HttpErr = http.StatusForbidden
		res.Messages = append(res.Messages, ERR_UserExists)
	}

	// now error till this point, let's register the new user
	if c.HttpErr == 0 {

		if !x.RSetSingle("user", c.User, x.CryptString(c.Token), 0) {

			c.HttpErr = http.StatusInternalServerError
			res.Messages = append(res.Messages, ERR_UserSetupFailed)
		} else {
			c.HttpErr = http.StatusCreated
			m := fmt.Sprintf("OK: User '%s' created.", c.User)
			res.Messages = append(res.Messages, m)
		}
	}

	returnWebResponse(w, c.HttpErr, &res)
}

func adminHandler(w http.ResponseWriter, r *http.Request) {

	res := x.NewWebResponse()
	c := x.NewCacheData()

	adminToken := r.URL.Query().Get("token")
	entity := r.URL.Query().Get("entity")
	action := r.URL.Query().Get("action")

	if adminToken == "" {
		c.HttpErr = http.StatusInternalServerError
		res.Messages = append(res.Messages, "ERR: Missing 'token' parameter")
	}

	if entity == "user" {
		if action == "purge" {

		} else {
			c.HttpErr = http.StatusInternalServerError
			res.Messages = append(res.Messages, "ERR: Illegal 'action' parameter")
		}

	} else if entity == "env" {
		if action == "create" {

		} else if action == "unlock" {

		} else if action == "terminate" {

		} else {
			c.HttpErr = http.StatusInternalServerError
			res.Messages = append(res.Messages, "ERR: Illegal 'action' parameter")
		}

	} else if entity == "host" {
		if action == "unlock" {

		} else if action == "terminate" {

		} else {
			c.HttpErr = http.StatusInternalServerError
			res.Messages = append(res.Messages, "ERR: Illegal 'action' parameter")
		}

	} else {
		c.HttpErr = http.StatusInternalServerError
		res.Messages = append(res.Messages, "ERR: Illegal 'name' parameter")
	}

	returnWebResponse(w, c.HttpErr, &res)

}

func returnWebResponse(w http.ResponseWriter, httpErr int, retData *x.WebResponseType) {

	// set 200 instead of 0 on http status
	if httpErr == 0 {
		httpErr = http.StatusOK
	}

	// set 'success' in JSON response according to http status code
	if httpErr < 400 {
		retData.Success = true
	} else {
		retData.Success = false
	}

	byteData, err := json.MarshalIndent(retData, "", "    ")
	if err != nil {
		http.Error(w, ERR_JsonConvertData, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", C_RespHeader)
	w.WriteHeader(httpErr)

	/* trunk-ignore(golangci-lint/errcheck) */
	w.Write(byteData)
}

func main() {

	x.RConn = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       1,
	})

	errDb := x.RConn.Ping().Err()
	if errDb == nil {
		fmt.Println("✅ Redis check OK")
	} else {
		fmt.Println("❌ Redis is not available, exitting...")
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

	if C_TLS {
		x.ServeTLS(r)
	} else {
		err := http.ListenAndServe("0.0.0.0:3000", r)
		if err != nil {
			log.Fatal("ListenAndServe: ", err)
		}
	}
}
