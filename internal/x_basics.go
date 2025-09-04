package x

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"time"
)

// Wants: n/a
//
// Returns: empty WebResponse structure
// func NewWebResponse() WebResponse {

// 	var r WebResponse

// 	r.Type = ""
// 	r.Name = ""
// 	r.Parent = ""
// 	r.State = ""
// 	r.LastDay = ""
// 	r.User = ""
// 	r.Success = false
// 	r.Messages = []string{}

// 	return r
// }

// Wants: n/a
//
// Returns: empty LockData structure
// func NewLockData() LockData {

// 	var c LockData

// 	c.Type = ""
// 	c.Name = ""
// 	c.Parent = ""
// 	c.State = ""
// 	c.LastDay = ""
// 	c.User = ""
// 	c.Token = ""
// 	c.HttpErr = 0

// 	return c
// }

// DEPRECATED: This function is no longer used. Password hashing is now handled by HashPassword() in x_password.go
// which uses bcrypt. This function remains for backward compatibility only.
// func CryptString(plain string) string {
//
// 	const preSalt string = "68947b1f416c3a5655e1ff9e7c7935f6"
// 	const postSalt string = "5f09dd9c81596ea3cc93ce0df58e26d8"
//
// 	h := sha1.New()
// 	h.Write([]byte(preSalt + plain + postSalt))
// 	sha1Hash := h.Sum(nil)
// 	hexString := fmt.Sprintf("%x", sha1Hash)
// 	return hexString
// }

// Wants: a string containing a date in YYYYMMDD format
//
// Returns: time from now till the last second of YYYYMMMDD specified date
func GetTimeFromNow(yyyymmdd string) time.Duration {

	// Parse the YYYYMMDD formatted datetime string into a time.Time object
	date, err := time.Parse("20060102", yyyymmdd)
	if err != nil {
		log.Fatal(err)
	}

	// Add one day to the parsed date and set the time to midnight
	nextDay := date.AddDate(0, 0, 1)
	firstSecondNextDay := time.Date(nextDay.Year(), nextDay.Month(), nextDay.Day(), 0, 0, 0, 0, nextDay.Location())

	// Calculate the duration from now to the first second of the next day
	duration := time.Until(firstSecondNextDay)

	return duration
}

// Wants: a string with YYYYMMDD date
//
// Returns: `true` if that date is a valid date
func IsValidDate(dateParam string) bool {

	if len(dateParam) != 8 {
		if DEBUG {
			fmt.Println("bad date length")
		}
		return false // Bad date length
	}

	// Compile the regular expression
	regex := regexp.MustCompile(`^\d{8}$`)

	// Check if the input matches the pattern
	if !regex.MatchString(dateParam) {
		return false // Input doesn't match the pattern
	}

	// Define the layout for YYYYMMDD format
	layout := "20060102"

	// Parse the dateString into a time.Time object
	parsedDate, err := time.Parse(layout, dateParam)
	if err != nil {
		return false // Parsing failed
	}

	// Check if the parsed date matches the input
	if parsedDate.Format(layout) != dateParam {
		return false
	}

	return true
}

// Wants: `env|host:name` key in Redis, expire date in YYYYMMDD
//
// Returns: `true` on success
func ExpireEntity(entity string, expireAt string) bool {

	if !IsValidDate(expireAt) { //
		return false
	}

	if !RSetExpire(entity, GetTimeFromNow(expireAt)) {
		return false
	}

	return true
}

// Wants: a string which should be C_TYPE_ENV or C_TYPE_HOST
//
// Returns: Specific RichErrorStatus
func ValidateType(t string) *RichErrorStatus {

	r := new(RichErrorStatus)

	switch t {
	case C_TYPE_ENV, C_TYPE_HOST:
		r.IsError = false
		r.HttpErrCode = C_HTTP_OK
		r.ErrorMessage = ""
		return r
	case "":
		r.IsError = true
		r.HttpErrCode = http.StatusBadRequest
		r.ErrorMessage = ERR_NoTypeSpecified
		return r
	default:
		r.IsError = true
		r.HttpErrCode = http.StatusBadRequest
		r.ErrorMessage = ERR_WrongTypeSpecified
		return r
	}
}

// Wants: username
//
// Returns: `true` if user exists
func IsExistingUser(userName string) bool {

	if DEBUG {
		fmt.Println("IsExistingUser <<", userName)
	}

	_, err := RConn.HGet("user", userName).Result()

	if err == nil {
		if DEBUG {
			fmt.Println("IsExistingUser >>", true)
		}
		return true
	} else {
		if DEBUG {
			fmt.Println("IsExistingUser >>", false)
		}
		return false
	}
}

// Wants: username, usertoken
//
// Returns: `true` if user is valid (password is matching)
func IsValidUser(userName string, userToken string) bool {
	if DEBUG {
		fmt.Println("IsValidUser <<", userName)
	}

	redisPwd, err := RConn.HGet("user", userName).Result()

	if err == nil && len(redisPwd) > 0 {
		// found user & password
		if CheckPassword(userToken, redisPwd) {
			// If using old hash format, upgrade to bcrypt
			if NeedsUpgrade(redisPwd) {
				if newHash, err := HashPassword(userToken); err == nil {
					RSetSingle("user", userName, newHash, 0)
				}
			}
			if DEBUG {
				fmt.Println("IsValidUser >>", true)
			}
			return true
		} else {
			if DEBUG {
				fmt.Println("IsValidUser >>", false)
			}
			return false
		}
	} else {
		if DEBUG {
			fmt.Println("IsValidUser >> NO_USER")
		}
		return false
	}
}

// Wants: hostname
//
// Returns: first tag from full hostname by separator character list
func GetEnvFromHost(hostName string) string {

	separators := []string{"-", "_", "/", ".", "|"}
	pos := 0

	if DEBUG {
		fmt.Println("GetEnvFromHost <<", hostName)
	}

	for i, char := range hostName {
		for _, separator := range separators {
			if string(char) == separator && pos == 0 {
				pos = i
				break
			}
		}
	}

	ret := hostName[:pos]
	if DEBUG {
		fmt.Println("GetEnvFromHost >>", ret)
	}

	return ret
}

func EnvCreate(envName string) bool {

	c := new(LockData)
	c.Type = C_TYPE_ENV
	c.Name = envName
	c.State = C_STATE_VALID
	return RLockSetter(c)
}

func EnvMaintenance(envName string) bool {

	c := new(LockData)
	c.Type = C_TYPE_ENV
	c.Name = envName
	c.State = C_STATE_MAINTENANCE
	return RLockSetter(c)
}

func EnvTerminate(envName string) bool {

	c := RLockGetter(C_TYPE_ENV + ":" + envName)
	c.Type = C_TYPE_ENV
	c.Name = envName
	c.State = C_STATE_TERMINATED
	c.Parent = "n/a"
	c.User = C_ADMIN
	return RLockSetter(c)
}

// Wants: environment name
//
// Returns: Specific RichErrorStatus
func EnvLockStatus(envName string) *RichErrorStatus {

	r := new(RichErrorStatus)
	ld := RLockGetter(C_TYPE_ENV + ":" + envName)

	switch {
	case ld.HttpErr == http.StatusNoContent:
		r.IsError = true
		r.HttpErrCode = http.StatusNoContent

	case ld.State == C_STATE_LOCKED:
		// no such env, returning locked state
		r.IsError = false
		r.HttpErrCode = http.StatusLocked

	case ld.State == C_STATE_VALID:
		r.IsError = false
		r.HttpErrCode = http.StatusAccepted

	case ld.State == "":
		r.IsError = false
		r.HttpErrCode = http.StatusAccepted

	default:
		r.IsError = true
		r.HttpErrCode = http.StatusInternalServerError
	}

	return r
}

// Wants: filled LockData
//
// Returns: `true` if everything went fine
func EnvLock(c *LockData, res *WebResponse) bool {

	db := RLockGetter(C_TYPE_ENV + ":" + c.Name)

	// normal users can modify only their own records
	if c.User != C_ADMIN {
		if c.User != db.User && len(db.User) > 0 {
			res.Messages = append(res.Messages, ERR_LockedByAnotherUser)
			c.HttpErr = http.StatusForbidden
			return false
		}
	}

	c.Parent = "n/a"
	c.State = C_STATE_LOCKED
	res.Messages = append(res.Messages, OK_EnvLocked)

	if RLockSetter(c) { // OK
		c.HttpErr = http.StatusOK
		return true

	} else { // ERROR
		res.Messages = append(res.Messages, ERR_EnvLockFail)
		c.HttpErr = http.StatusForbidden
		return false
	}
}

func EnvUnlock(envName string) bool {

	c := new(LockData)
	c.Type = C_TYPE_ENV
	c.Name = envName
	c.State = C_STATE_VALID
	return RLockSetter(c)
}

func IsEnvContainsHosts(envName string) bool {

	if len(RGetHostsInEnv(envName)) > 0 {
		return true
	} else {
		return false
	}
}

func IsHostLocked(hostName string) bool {

	c := RLockGetter(C_TYPE_HOST + ":" + hostName)
	if c.State == C_STATE_LOCKED {
		return false
	} else {
		return true
	}
}

// Wants: filled LockData
//
// Returns: `true` if everything went fine
func HostLock(c *LockData, res *WebResponse) bool {

	c.Parent = GetEnvFromHost(c.Name)             // parent env
	pl := EnvLockStatus(c.Parent)                 // parent locking status
	db := RLockGetter(C_TYPE_HOST + ":" + c.Name) // host locking status

	if pl.HttpErrCode == http.StatusLocked { // parent env has been locked
		res.Messages = append(res.Messages, ERR_ParentEnvLockFail)
		c.HttpErr = http.StatusForbidden
		return false
	}

	if pl.HttpErrCode == http.StatusNoContent { // parent env not defined
		res.Messages = append(res.Messages, ERR_ParentEnvNil)
		c.HttpErr = http.StatusForbidden
		return false
	}

	if c.State == C_STATE_LOCKED {
		res.Messages = append(res.Messages, ERR_HostLockFail)
		c.HttpErr = http.StatusForbidden
		return false
	}

	// normal users can modify only their own records
	if c.User != C_ADMIN {
		if c.User != db.User && len(db.User) > 0 {
			res.Messages = append(res.Messages, ERR_LockedByAnotherUser)
			c.HttpErr = http.StatusForbidden
			return false
		}
	}

	c.State = C_STATE_LOCKED

	if RLockSetter(c) { // OK
		res.Messages = append(res.Messages, OK_HostLocked)
		c.HttpErr = http.StatusOK
		return true

	} else { // ERROR
		res.Messages = append(res.Messages, ERR_HostLockFail)
		c.HttpErr = http.StatusForbidden
		return false
	}
}

func HostUnlock(hostName string) bool {

	return REntityDelete(C_TYPE_HOST, hostName)
}
