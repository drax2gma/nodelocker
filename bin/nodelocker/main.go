// main.go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"text/template"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/go-redis/redis"

	x "github.com/drax2gma/nodelocker/internal"
)

func jsonStatHandler(w http.ResponseWriter, r *http.Request) {

	stats := new(x.Stats)
	x.RFillJsonStats(stats)

	w.Header().Set("Content-Type", x.C_RespHeader)
	w.WriteHeader(http.StatusOK)

	byteData, err := json.MarshalIndent(stats, "", "    ")
	if err != nil {
		http.Error(w, x.ERR_JsonConvertData, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", x.C_RespHeader)
	/* trunk-ignore(golangci-lint/errcheck) */
	w.Write(byteData)
}

func webStatHandler(w http.ResponseWriter, r *http.Request) {

	stats := new(x.Stats)
	x.RFillJsonStats(stats)

	tmpl := template.Must(template.New("index").Parse(`
	<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.5">
		<title>Nodelocker overview</title>
		<style>
			body {
				font-family: Arial, sans-serif;
				background-color: #f4f4f4;
				margin: 0;
				padding: 0;
			}
			.container {
				max-width: 800px;
				margin: 20px auto;
				padding: 20px;
				background-color: #fff;
				border-radius: 8px;
				box-shadow: 0 0 10px rgba(0, 0, 0, 0.1);
			}
			h1 {
				text-align: center;
			}
			ul {
				list-style: none;
				padding: 0;
			}
			li {
				margin-bottom: 5px;
			}
			.label {
				font-weight: bold;
			}
		</style>
	</head>
	<body>
		<div class="container">
			<h1>🏗️ Environment and host overview 🏗️</h1>
			<br><hr><br>
			<div id="stats">
				<ul>
					<li><span class="label">Valid environments:</span>
						<ul>{{range .ValidEnvs}}
							<li><span class="value">✅ {{.}}</span></li>
						{{end}}</ul>
					</li>
					<br>
					<li><span class="label">Locked environments:</span>
						<ul>{{range .LockedEnvs}}
							<li><span class="value">🔒 {{.}}</span></li>
						{{end}}</ul>
					</li>
					<br>
					<li><span class="label">Environments in maintenance mode:</span>
						<ul>{{range .MaintEnvs}}
							<li><span class="value">🚧 {{.}}</span></li>
						{{end}}</ul>
					</li>
					<br>
					<li><span class="label">Terminated, unusable environments:</span>
						<ul>{{range .TermdEnvs}}
							<li><span class="value">❌ {{.}}</span></li>
						{{end}}</ul>
					</li>
					<br><hr><br>
					<li><span class="label"> Locked hosts:</span>
						<ul>{{range .LockedHosts}}
							<li><span class="value">🔒 {{.}}</span></li>
						{{end}}</ul>
					</li>
				</ul>
			</div>
		</div>
	</body>
	</html>
	`))

	if err := tmpl.Execute(w, stats); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func lockHandler(w http.ResponseWriter, r *http.Request) {

	res := new(x.WebResponse)
	c := new(x.LockData)

	c.Type = r.URL.Query().Get("type")
	c.Name = r.URL.Query().Get("name")
	c.LastDay = r.URL.Query().Get("lastday")
	c.User = r.URL.Query().Get("user")
	c.Token = r.URL.Query().Get("token")

	// Check if init sequence has been made when starting anything as normal user
	if c.User != x.C_ADMIN {
		if !x.IsExistingUser(x.C_ADMIN) {

			c.HttpErr = http.StatusLocked
			res.Messages = append(res.Messages, x.ERR_NoAdminPresent)
		}
	}

	// check 'type' defined in GET request
	t := x.ValidateType(c.Type)
	if t.IsError {

		c.HttpErr = t.HttpErrCode
		res.Messages = append(res.Messages, t.ErrorMessage)
	}

	// no 'name' defined in GET request
	if c.Name == "" {

		c.HttpErr = http.StatusBadRequest
		res.Messages = append(res.Messages, x.ERR_NoNameSpecified)
	}

	// trying to lock an env that contain locked host(s)
	if c.Type == x.C_TYPE_ENV && x.IsEnvContainsHosts(c.Name) {

		c.HttpErr = http.StatusForbidden
		res.Messages = append(res.Messages, x.ERR_LockedHostsInEnv)
	}

	// Is given LASTDAY is a valid date?
	if !x.IsValidDate(c.LastDay) {

		c.HttpErr = http.StatusBadRequest
		res.Messages = append(res.Messages, x.ERR_InvalidDateSpecified)
	}

	// Is given user valid against DB user? Pwd checking too.
	if !x.IsValidUser(c.User, c.Token) {

		c.HttpErr = http.StatusForbidden
		res.Messages = append(res.Messages, x.ERR_IllegalUser)
	}

	// on C_HTTP_OK lock
	if c.HttpErr == x.C_HTTP_OK {

		switch {
		case c.Type == x.C_TYPE_ENV:
			x.EnvLock(c, res)
		case c.Type == x.C_TYPE_HOST:
			x.HostLock(c, res)
		default:
		}

		if !x.ExpireEntity(c.Type+":"+c.Name, c.LastDay) {
			c.HttpErr = http.StatusBadRequest
			res.Messages = append(res.Messages, x.ERR_InvalidDateSpecified)
		}
	}

	returnWebResponse(w, c.HttpErr, res)
}

func unlockHandler(w http.ResponseWriter, r *http.Request) {

	res := new(x.WebResponse)
	c := new(x.LockData)

	c.Type = r.URL.Query().Get("type")
	c.Name = r.URL.Query().Get("name")
	c.User = r.URL.Query().Get("user")
	c.Token = r.URL.Query().Get("token")

	// Check if init sequence has been made when starting anything as normal user
	if c.User != x.C_ADMIN {
		if !x.IsExistingUser(x.C_ADMIN) {

			c.HttpErr = http.StatusLocked
			res.Messages = append(res.Messages, x.ERR_NoAdminPresent)
		}
	}

	// check 'type' defined in GET request
	t := x.ValidateType(c.Type)
	if t.IsError {

		c.HttpErr = http.StatusBadRequest
		res.Messages = append(res.Messages, x.ERR_WrongTypeSpecified)
	}

	// Check for missing entity name
	if c.Name == "" {

		c.HttpErr = http.StatusBadRequest
		res.Messages = append(res.Messages, x.ERR_NoNameSpecified)
	}

	if !x.IsValidUser(c.User, c.Token) {

		c.HttpErr = http.StatusForbidden
		res.Messages = append(res.Messages, x.ERR_IllegalUser)
	}

	if !x.REntityDelete(c.Type, c.Name) {

		c.HttpErr = http.StatusForbidden
		res.Messages = append(res.Messages, "ERR: EntityDelete failed !!!")

	} else {

		c.HttpErr = http.StatusOK
		res.Messages = append(res.Messages, "")
	}

	returnWebResponse(w, c.HttpErr, res)
}

func regHandler(w http.ResponseWriter, r *http.Request) {

	res := new(x.WebResponse)
	c := new(x.LockData)

	c.User = r.URL.Query().Get("user")
	c.Token = r.URL.Query().Get("token")

	// Check if init sequence has been made when starting anything as normal user
	if c.User != x.C_ADMIN {
		if !x.IsExistingUser(x.C_ADMIN) {

			c.HttpErr = http.StatusLocked
			res.Messages = append(res.Messages, x.ERR_NoAdminPresent)
		}
	}

	// no 'user' defined in GET request
	if c.User == "" {

		c.HttpErr = http.StatusBadRequest
		res.Messages = append(res.Messages, x.ERR_NoUserSpecified)
	} else {
		res.User = c.User
	}

	// no 'token' defined in GET request
	if c.Token == "" {

		c.HttpErr = http.StatusBadRequest
		res.Messages = append(res.Messages, x.ERR_NoTokenSpecified)
	}

	if x.IsExistingUser(c.User) {

		c.HttpErr = http.StatusForbidden
		res.Messages = append(res.Messages, x.ERR_UserExists)
	}

	// now error till this point, let's register the new user
	if c.HttpErr == x.C_HTTP_OK {

		hashedPassword, err := x.HashPassword(c.Token)
		if err != nil {
			c.HttpErr = http.StatusInternalServerError
			res.Messages = append(res.Messages, "Error hashing password")
			return
		}

		if !x.RSetSingle("user", c.User, hashedPassword, 0) {

			c.HttpErr = http.StatusInternalServerError
			res.Messages = append(res.Messages, x.ERR_UserSetupFailed)
		} else {
			c.HttpErr = http.StatusCreated
			m := fmt.Sprintf("OK: User '%s' created.", c.User)
			res.Messages = append(res.Messages, m)
		}
	}

	returnWebResponse(w, c.HttpErr, res)
}

func adminHandler(w http.ResponseWriter, r *http.Request) {

	res := new(x.WebResponse)
	c := new(x.LockData)

	action := r.URL.Query().Get("action")
	c.Name = r.URL.Query().Get("name")
	adminToken := r.URL.Query().Get("token")

	if adminToken == "" || !x.IsValidUser(x.C_ADMIN, adminToken) {
		c.HttpErr = http.StatusForbidden
		res.Messages = append(res.Messages, x.ERR_IllegalUser)

	} else if action == "user-purge" { // Purge a user which probably forgot their password

		if x.REntityDelete("user", c.Name) {
			c.HttpErr = http.StatusOK
			res.Messages = append(res.Messages, x.OK_UserPurged)
		} else {
			c.HttpErr = http.StatusInternalServerError
			res.Messages = append(res.Messages, x.ERR_CannotDeleteUser)
		}

	} else if action == "env-create" { // Add a new environment

		if x.EnvCreate(c.Name) {
			c.HttpErr = http.StatusOK
			res.Messages = append(res.Messages, x.OK_EnvCreated)
		} else {
			c.HttpErr = http.StatusForbidden
			res.Messages = append(res.Messages, x.ERR_EnvCreationFail)
		}

	} else if action == "env-unlock" { // Unlock an env from maintenance or terminate state

		if x.EnvUnlock(c.Name) {
			c.HttpErr = http.StatusOK
			res.Messages = append(res.Messages, x.OK_EnvUnlocked)
		} else {
			c.HttpErr = http.StatusForbidden
			res.Messages = append(res.Messages, x.ERR_EnvUnlockFail)
		}

	} else if action == "env-maintenance" { // Setup an env for maintenance

		if x.EnvMaintenance(c.Name) {
			c.HttpErr = http.StatusOK
			c.State = x.C_STATE_MAINTENANCE
			res.Messages = append(res.Messages, x.OK_EnvSetToMaintenance)
		} else {
			c.HttpErr = http.StatusForbidden
			res.Messages = append(res.Messages, x.ERR_EnvSetMaintFailFail)
		}

	} else if action == "env-terminate" { // Lock an env indefinately

		if x.EnvTerminate(c.Name) {
			c.HttpErr = http.StatusOK
			c.State = x.C_STATE_TERMINATED
			res.Messages = append(res.Messages, x.OK_EnvSetToTerminate)
		} else {
			c.HttpErr = http.StatusForbidden
			res.Messages = append(res.Messages, x.ERR_EnvSetTermFail)
		}

	} else if action == "host-unlock" { // Unlock a stuck, locked host

		if x.HostUnlock(c.Name) {
			c.HttpErr = http.StatusOK
			res.Messages = append(res.Messages, x.OK_HostUnlocked)
		} else {
			c.HttpErr = http.StatusForbidden
			res.Messages = append(res.Messages, x.ERR_HostUnlockFail)
		}

	} else {
		c.HttpErr = http.StatusInternalServerError
		res.Messages = append(res.Messages, "ERR: Illegal 'action' parameter")
	}

	returnWebResponse(w, c.HttpErr, res)
}

func returnWebResponse(w http.ResponseWriter, httpErr int, retData *x.WebResponse) {

	// set 200 instead of C_HTTP_OK on http status
	if httpErr == x.C_HTTP_OK {
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
		http.Error(w, x.ERR_JsonConvertData, http.StatusInternalServerError)
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

	r.Get("/status/json", jsonStatHandler)
	r.Get("/status/web", webStatHandler)
	r.Get("/lock", lockHandler)
	r.Get("/unlock", unlockHandler)
	r.Get("/register", regHandler)
	r.Get("/admin", adminHandler)

	http.Handle("/", r)

	if x.C_TLS_ENABLED {
		x.ServeTLS(r)
	} else {
		err := http.ListenAndServe("0.0.0.0:3000", r)
		if err != nil {
			log.Fatal("ListenAndServe: ", err)
		}
	}
}
