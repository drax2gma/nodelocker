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
	ERR_JsonConvertData      string = "ERR: Error converting LockData to JSON."
	ERR_NoNameSpecified      string = "ERR: No 'name' parameter specified."
	ERR_NoTypeSpecified      string = "ERR: No 'type' parameter specified."
	ERR_NoUserSpecified      string = "ERR: No 'user' parameter specified."
	ERR_NoTokenSpecified     string = "ERR: No 'token' parameter specified."
	ERR_WrongTypeSpecified   string = "ERR: Wrong 'type' specified, must be 'env' or 'host'."
	ERR_IllegalUser          string = "ERR: Illegal user."
	ERR_CannotDeleteUser     string = "ERR: Cannot delete user."
	ERR_UserExists           string = "ERR: User already exists."
	ERR_UserSetupFailed      string = "ERR: User setup failed."
	ERR_EnvLockFail          string = "ERR: Environment lock unsuccesful."
	ERR_EnvCreationFail      string = "ERR: Creating a new enviromnent failed."
	ERR_EnvUnlockFail        string = "ERR: Environment unlock failed."
	ERR_EnvSetMaintFailFail  string = "ERR: Environment maintenance set failed."
	ERR_EnvSetTermFail       string = "ERR: Environment termination failed."
	ERR_HostLockFail         string = "ERR: Host lock unsuccesful."
	ERR_HostUnlockFail       string = "ERR: Host unlock failed."
	ERR_InvalidDateSpecified string = "ERR: Invalid 'lastday' specified, format is: YYYYMMDD."
	ERR_NoAdminPresent       string = "ERR: No 'admin' user present, cannot continue."

	OK_UserPurged          string = "OK: User purged."
	OK_EnvCreated          string = "OK: Environment created."
	OK_EnvUnlocked         string = "OK: Environment unlocked."
	OK_EnvSetToMaintenance string = "OK: Environment is in maintenance mode now."
	OK_EnvSetToTerminate   string = "OK: Environment terminated."
	OK_HostUnlocked        string = "OK: Host has been unlocked succesfully."

	C_TLS bool = true // serve TLS with self-signed cert?

	C_HTTP_OK = 0
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
	if !x.IsValidEntityType(c.Type) {

		c.HttpErr = http.StatusBadRequest
		res.Messages = append(res.Messages, ERR_WrongTypeSpecified)
	}

	// no 'name' defined in GET request
	if c.Name == "" {

		c.HttpErr = http.StatusBadRequest
		res.Messages = append(res.Messages, ERR_NoNameSpecified)
	}

	if c.HttpErr == C_HTTP_OK { // GET params were good

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

		w.Header().Set("Content-Type", x.C_RespHeader)
		w.WriteHeader(c.HttpErr)
	}

	byteData, err := json.MarshalIndent(res, "", "    ")
	if err != nil {
		http.Error(w, ERR_JsonConvertData, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", x.C_RespHeader)
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
	if c.User != x.C_ADMIN {
		if !x.IsExistingUser(x.C_ADMIN) {

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
	if !x.IsValidUser(c.User, c.Token) {

		c.HttpErr = http.StatusForbidden
		res.Messages = append(res.Messages, ERR_IllegalUser)
	}

	if !x.RSetLockData(&c) {

		c.HttpErr = http.StatusInternalServerError

		if c.Type == x.C_TYPE_ENV {
			res.Messages = append(res.Messages, ERR_EnvLockFail)
		} else if c.Type == x.C_TYPE_HOST {
			res.Messages = append(res.Messages, ERR_HostLockFail)
		} else {
			res.Messages = append(res.Messages, ERR_InvalidDateSpecified)
		}
	}

	if c.HttpErr == C_HTTP_OK { // GET params were good

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

		w.Header().Set("Content-Type", x.C_RespHeader)
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
		if !x.IsExistingUser(x.C_ADMIN) {

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

	if !x.IsValidUser(c.User, c.Token) {

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
		if !x.IsExistingUser(x.C_ADMIN) {

			c.HttpErr = http.StatusLocked
			res.Messages = append(res.Messages, ERR_NoAdminPresent)
		}
	}

	// no 'user' defined in GET request
	if c.User == "" {

		c.HttpErr = http.StatusBadRequest
		res.Messages = append(res.Messages, ERR_NoUserSpecified)
	} else {
		res.User = c.User
	}

	// no 'token' defined in GET request
	if c.Token == "" {

		c.HttpErr = http.StatusBadRequest
		res.Messages = append(res.Messages, ERR_NoTokenSpecified)
	}

	if x.IsExistingUser(c.User) {

		c.HttpErr = http.StatusForbidden
		res.Messages = append(res.Messages, ERR_UserExists)
	}

	// now error till this point, let's register the new user
	if c.HttpErr == C_HTTP_OK {

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

	action := r.URL.Query().Get("action")
	name := r.URL.Query().Get("name")
	adminToken := r.URL.Query().Get("token")

	if adminToken == "" || !x.IsValidUser(x.C_ADMIN, adminToken) {
		c.HttpErr = http.StatusForbidden
		res.Messages = append(res.Messages, ERR_IllegalUser)

	} else if action == "user-purge" { // Purge a user which probably forgot their password

		if x.REntityDelete("user", name) {
			c.HttpErr = http.StatusOK
			res.Messages = append(res.Messages, OK_UserPurged)
		} else {
			c.HttpErr = http.StatusInternalServerError
			res.Messages = append(res.Messages, ERR_CannotDeleteUser)
		}

	} else if action == "env-create" { // Add a new environment

		if x.CreateEnv(name) {
			c.HttpErr = http.StatusOK
			res.Messages = append(res.Messages, OK_EnvCreated)
		} else {
			c.HttpErr = http.StatusForbidden
			res.Messages = append(res.Messages, ERR_EnvCreationFail)
		}

	} else if action == "env-unlock" { // Unlock an env from maintenance or terminate state

		if x.UnlockEnv(name) {
			c.HttpErr = http.StatusOK
			res.Messages = append(res.Messages, OK_EnvUnlocked)
		} else {
			c.HttpErr = http.StatusForbidden
			res.Messages = append(res.Messages, ERR_EnvUnlockFail)
		}

	} else if action == "env-maintenance" { // Setup an env for maintenance

		if x.MaintenanceEnv(name) {
			c.HttpErr = http.StatusOK
			res.Messages = append(res.Messages, OK_EnvSetToMaintenance)
		} else {
			c.HttpErr = http.StatusForbidden
			res.Messages = append(res.Messages, ERR_EnvSetMaintFailFail)
		}

	} else if action == "env-terminate" { // Lock an env indefinately

		if x.TerminateEnv(name) {
			c.HttpErr = http.StatusOK
			res.Messages = append(res.Messages, OK_EnvSetToTerminate)
		} else {
			c.HttpErr = http.StatusForbidden
			res.Messages = append(res.Messages, ERR_EnvSetTermFail)
		}

	} else if action == "host-unlock" { // Unlock a stuck, locked host

		if x.UnlockHost(name) {
			c.HttpErr = http.StatusOK
			res.Messages = append(res.Messages, OK_HostUnlocked)
		} else {
			c.HttpErr = http.StatusForbidden
			res.Messages = append(res.Messages, ERR_HostUnlockFail)
		}

	} else {
		c.HttpErr = http.StatusInternalServerError
		res.Messages = append(res.Messages, "ERR: Illegal 'action' parameter")
	}

	returnWebResponse(w, c.HttpErr, &res)
}

func returnWebResponse(w http.ResponseWriter, httpErr int, retData *x.WebResponseType) {

	// set 200 instead of C_HTTP_OK on http status
	if httpErr == C_HTTP_OK {
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

	w.Header().Set("Content-Type", x.C_RespHeader)
	w.WriteHeader(httpErr)

	/* trunk-ignore(golangci-lint/errcheck) */
	w.Write(byteData)
}

func main() {

	x.RConn = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	errDb := x.RConn.Ping().Err()
	if errDb == nil {
		fmt.Println(x.C_SUCCESS + " Redis check OK")
	} else {
		fmt.Println(x.C_FAILED + " Redis is not available, exitting...")
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
