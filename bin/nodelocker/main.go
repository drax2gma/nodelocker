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

	var webResponse x.WebResponseDataType
	var httpErrorCode int = 0

	x.ResetWebResponse(&webResponse)

	x.CacheData.Type = r.URL.Query().Get("type") // type of entity, 'env' or 'host'
	x.CacheData.Name = r.URL.Query().Get("name") // name of entity

	// no 'type' defined in GET request
	if x.CacheData.Type == "" {

		httpErrorCode = http.StatusBadRequest
		webResponse.Messages = append(webResponse.Messages, ERR_NoTypeSpecified)
	}

	// bad 'type' defined in GET request
	if x.CacheData.Type != "env" && x.CacheData.Type != "host" {

		httpErrorCode = http.StatusBadRequest
		webResponse.Messages = append(webResponse.Messages, ERR_WrongTypeSpecified)
	}

	// no 'name' defined in GET request
	if x.CacheData.Name == "" {

		httpErrorCode = http.StatusBadRequest
		webResponse.Messages = append(webResponse.Messages, ERR_NoNameSpecified)
	}

	if httpErrorCode == 0 { // GET params were good

		if x.RGetLockData() {
			// got some data on entity from lock database
			webResponse.State = x.CacheData.State
			webResponse.User = x.CacheData.User
		} else {
			// no lock data on entity
			webResponse.State = "unlocked"
			webResponse.User = ""
		}

		webResponse.Success = true
		webResponse.Messages = []string{"valid_response"}
		webResponse.Type = x.CacheData.Type
		webResponse.Name = x.CacheData.Name

	} else { // GET params were problametic a bit

		w.Header().Set("Content-Type", C_RespHeader)
		w.WriteHeader(httpErrorCode)
	}

	byteData, err := json.MarshalIndent(webResponse, "", "    ")
	if err != nil {
		http.Error(w, ERR_JsonConvertData, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", C_RespHeader)
	/* trunk-ignore(golangci-lint/errcheck) */
	w.Write(byteData)
}

func lockHandler(w http.ResponseWriter, r *http.Request) {

	var webResponse x.WebResponseDataType
	var httpErrorCode int = 0

	x.ResetWebResponse(&webResponse)

	x.CacheData.Type = r.URL.Query().Get("type")
	x.CacheData.Name = r.URL.Query().Get("name")
	x.CacheData.LastDay = r.URL.Query().Get("lastday")
	x.CacheData.User = r.URL.Query().Get("user")
	x.CacheData.Token = r.URL.Query().Get("token")

	// no 'type' defined in GET request
	if x.CacheData.Type == "" {

		httpErrorCode = http.StatusBadRequest
		webResponse.Messages = append(webResponse.Messages, ERR_NoTypeSpecified)
	}

	// bad 'type' defined in GET request
	if !x.IsValidEntityType(x.CacheData.Type) {

		httpErrorCode = http.StatusBadRequest
		webResponse.Messages = append(webResponse.Messages, ERR_WrongTypeSpecified)
	}

	// no 'name' defined in GET request
	if x.CacheData.Name == "" {

		httpErrorCode = http.StatusBadRequest
		webResponse.Messages = append(webResponse.Messages, ERR_NoNameSpecified)
	}

	// Is given LASTDAY is a valid date?
	if !x.IsValidDate(x.CacheData.LastDay) {

		httpErrorCode = http.StatusBadRequest
		webResponse.Messages = append(webResponse.Messages, ERR_InvalidDateSpecified)
	}

	// Is given user valid against DB user? Pwd checking too.
	if x.RCheckUser(x.C_USER_Valid, x.CacheData.User, x.CacheData.Token) == x.C_USER_Invalid {

		httpErrorCode = http.StatusForbidden
		webResponse.Messages = append(webResponse.Messages, ERR_IllegalUser)
	}

	//
	if !x.RSetSingle(x.C_UseCacheData, "state", C_LOCK, x.GetTimeFromNow(x.CacheData.LastDay)) {

		httpErrorCode = http.StatusBadRequest
		webResponse.Messages = append(webResponse.Messages, ERR_EnvLockFail)
	}

	//
	if !x.RSetSingle(x.C_UseCacheData, "user", x.CacheData.User, x.GetTimeFromNow(x.CacheData.LastDay)) {

		httpErrorCode = http.StatusBadRequest
		webResponse.Messages = append(webResponse.Messages, ERR_EnvLockFail)
	}

	//
	if !x.RSetSingle(x.C_UseCacheData, "lastday", x.CacheData.LastDay, x.GetTimeFromNow(x.CacheData.LastDay)) {

		httpErrorCode = http.StatusBadRequest
		webResponse.Messages = append(webResponse.Messages, ERR_EnvLockFail)
	}

	if httpErrorCode == 0 { // GET params were good

		if x.RGetLockData() {
			// got some data on entity from lock database
			webResponse.State = x.CacheData.State
			webResponse.User = x.CacheData.User
		} else {
			// no lock data on entity
			webResponse.State = "unlocked"
			webResponse.User = ""
		}

		webResponse.Success = true
		webResponse.Messages = []string{"valid_response"}
		webResponse.Type = x.CacheData.Type
		webResponse.Name = x.CacheData.Name

	} else { // GET params were problametic a bit

		w.Header().Set("Content-Type", C_RespHeader)
		w.WriteHeader(httpErrorCode)
	}

	byteData, err := json.MarshalIndent(webResponse, "", "    ")
	if err != nil {
		http.Error(w, ERR_JsonConvertData, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", C_RespHeader)
	/* trunk-ignore(golangci-lint/errcheck) */
	w.Write(byteData)
}

func unlockHandler(w http.ResponseWriter, r *http.Request) {

	var httpResponse string

	entityName := r.URL.Query().Get("entity")
	entityType := r.URL.Query().Get("type")
	userName := r.URL.Query().Get("user")
	userToken := r.URL.Query().Get("token")

	if entityName == "" {
		// only one parameter allowed, not host and env in the same time
		fmt.Println("+host+env")
		http.Error(w, "ERR: No 'entity' (host or env name) specified.", http.StatusBadRequest)
		return
	}

	if entityType == "" {
		// only one parameter allowed, not host and env in the same time
		fmt.Println("-host-env")
		http.Error(w, "ERR: No 'type' of entity specified.", http.StatusBadRequest)
		return
	}

	if x.RCheckUser(x.C_USER_Valid, userName, userToken) == x.C_USER_Invalid {
		http.Error(w, "Illegal user.", http.StatusForbidden)
		return
	}

	// UNLOCK function
	x.CacheData.Type = entityType
	x.CacheData.Name = entityName

	x.REntityDelete()
	httpResponse = fmt.Sprintf("'%s:%s' unlocked successfully by %s.", entityType, entityName, userName)

	w.Header().Set("Content-Type", C_RespHeader)
	/* trunk-ignore(golangci-lint/errcheck) */
	w.Write([]byte(string(httpResponse)))

}

func regHandler(w http.ResponseWriter, r *http.Request) {

	var statusResponse x.StatusRespType
	var httpErrorCode int = 0

	x.ResetStatusResponse(&statusResponse)

	userName := r.URL.Query().Get("user")
	userToken := r.URL.Query().Get("token")

	// First check if init has been made by adding an 'admin' user
	// if we need user operations
	if userName != "admin" {
		if x.RCheckUser(x.C_USER_Exists, "admin", "") != x.C_USER_Exists {

			httpErrorCode = http.StatusLocked
			statusResponse.Messages = append(statusResponse.Messages, ERR_NoAdminPresent)
		}
	}

	// no 'user' defined in GET request
	if userName == "" {

		httpErrorCode = http.StatusBadRequest
		statusResponse.Messages = append(statusResponse.Messages, ERR_NoUserSpecified)
	}

	// no 'token' defined in GET request
	if userToken == "" {

		httpErrorCode = http.StatusBadRequest
		statusResponse.Messages = append(statusResponse.Messages, ERR_NoTokenSpecified)
	}

	if x.RCheckUser(x.C_USER_Exists, userName, "") == x.C_USER_Exists {

		httpErrorCode = http.StatusForbidden
		statusResponse.Messages = append(statusResponse.Messages, ERR_UserExists)
	}

	// now error till this point, let's register the new user
	if httpErrorCode == 0 {

		if !x.RSetSingle("user", userName, x.CryptString(userToken), 0) {

			httpErrorCode = http.StatusInternalServerError
			statusResponse.Messages = append(statusResponse.Messages, ERR_UserSetupFailed)
		}
	}

	if httpErrorCode == 0 {

		statusResponse.Success = true
		statusResponse.Messages = []string{"OK: User created."}

	} else {

		w.Header().Set("Content-Type", C_RespHeader)
		w.WriteHeader(httpErrorCode)
	}

	byteData, err := json.MarshalIndent(statusResponse, "", "    ")
	if err != nil {
		http.Error(w, ERR_JsonConvertData, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", C_RespHeader)
	/* trunk-ignore(golangci-lint/errcheck) */
	w.Write(byteData)

}

func adminHandler(w http.ResponseWriter, r *http.Request) {

	var paramResult string

	adminToken := r.URL.Query().Get("token")
	entity := r.URL.Query().Get("entity")
	action := r.URL.Query().Get("action")

	if adminToken == "" {
		paramResult = "Missing 'token' parameter"
	}

	if entity == "user" {
		if action == "purge" {

		} else {
			paramResult = "ERR: Illegal 'action' parameter"
		}

	} else if entity == "env" {
		if action == "create" {

		} else if action == "unlock" {

		} else if action == "terminate" {

		} else {
			paramResult = "ERR: Illegal 'action' parameter"
		}

	} else if entity == "host" {
		if action == "unlock" {

		} else if action == "terminate" {

		} else {
			paramResult = "ERR: Illegal 'action' parameter"
		}

	} else {
		paramResult = "ERR: Illegal 'entity' parameter"
	}

	http.Error(w, paramResult, http.StatusBadRequest)

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
