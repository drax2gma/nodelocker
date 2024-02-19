package x

import (
	"crypto/sha1"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"time"
)

func NewWebResponse() WebResponseType {

	var r WebResponseType

	r.Type = ""
	r.Name = ""
	r.Parent = ""
	r.State = ""
	r.LastDay = ""
	r.User = ""
	r.Success = false
	r.Messages = []string{}

	return r
}

func NewCacheData() CacheDataType {

	var c CacheDataType

	c.Type = ""
	c.Name = ""
	c.Parent = ""
	c.State = ""
	c.LastDay = ""
	c.User = ""
	c.Token = ""
	c.HttpErr = 0

	return c
}

// Crypt a plain string into sha1 hash string
func CryptString(plain string) string {

	const preSalt string = "68947b1f416c3a5655e1ff9e7c7935f6"
	const postSalt string = "5f09dd9c81596ea3cc93ce0df58e26d8"

	h := sha1.New()
	h.Write([]byte(preSalt + plain + postSalt))
	sha1Hash := h.Sum(nil)
	hexString := fmt.Sprintf("%x", sha1Hash)
	return hexString
}

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

func IsValidDate(dateParam string) bool {

	if len(dateParam) != 8 {
		fmt.Println("bad date length")
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

// check 'type' defined in GET request
func CheckType(c *CacheDataType, res *WebResponseType) bool {

	if c.Type == C_TYPE_ENV || c.Type == C_TYPE_HOST {

		res.Type = c.Type
		return true

	} else if c.Type == "" {

		c.HttpErr = http.StatusBadRequest
		res.Messages = append(res.Messages, ERR_NoTypeSpecified)
		return false

	} else {

		c.HttpErr = http.StatusBadRequest
		res.Messages = append(res.Messages, ERR_WrongTypeSpecified)
		return false
	}

}

func IsExistingUser(userName string) bool {

	fmt.Println("IsExistingUser <<", userName)

	_, err := RConn.HGet("user", userName).Result()

	if err == nil {
		fmt.Println("IsExistingUser >>", true)
		return true
	} else {
		fmt.Println("IsExistingUser >>", false)
		return false
	}
}

func IsValidUser(userName string, userToken string) bool {

	fmt.Println("IsValidUser <<", userName)

	redisPwd, err := RConn.HGet("user", userName).Result()

	if err == nil && len(redisPwd) > 0 {
		// found user & password

		if redisPwd == CryptString(userToken) {
			fmt.Println("IsValidUser >>", true)
			return true
		} else {
			fmt.Println("IsValidUser >>", false)
			return false
		}
	} else {
		fmt.Println("IsValidUser >> NO_USER")
		return false
	}
}

// acquire the env from the hostname automagically
func GetEnvFromHost(hostName string) string {

	separators := []string{"-", "_", "/", ".", "|"}
	pos := 0

	fmt.Println("GetEnvFromHost <<", hostName)

	for i, char := range hostName {
		for _, separator := range separators {
			if string(char) == separator && pos == 0 {
				pos = i
				break
			}
		}
	}

	ret := hostName[:pos]
	fmt.Println("GetEnvFromHost >>", ret)

	return ret
}

func IsEnvLocked(envName string) bool {

	c := NewCacheData()
	c.Type = C_TYPE_ENV
	c.Name = envName

	if !RGetLockData(&c) {
		log.Fatal("ERR: IsEnvLocked")
	}

	switch {
	case c.State == C_STATE_VALID:
		return false
	case c.State == "":
		return false
	default:
		return true
	}
}

func EnvCreate(envName string) bool {

	// return RSetAddMember(C_ENV_LIST, envName)
	c := NewCacheData()
	c.Type = C_TYPE_ENV
	c.Name = envName
	c.State = C_STATE_VALID
	return RSetLockData(&c)
}

func EnvMaintenance(envName string) bool {

	c := NewCacheData()
	c.Type = C_TYPE_ENV
	c.Name = envName
	c.State = C_STATE_MAINTENANCE
	return RSetLockData(&c)
}

func EnvTerminate(envName string) bool {

	c := NewCacheData()
	c.Type = C_TYPE_ENV
	c.Name = envName

	p1 := RGetLockData(&c)

	c.State = C_STATE_TERMINATED

	p2 := RSetLockData(&c)

	return p1 && p2
}

func EnvUnlock(envName string) bool {

	c := NewCacheData()
	c.Type = C_TYPE_ENV
	c.Name = envName
	c.State = C_STATE_VALID
	return RSetLockData(&c)

}

func HostUnlock(hostName string) bool {

	return REntityDelete(C_TYPE_HOST, hostName)

}

func IsEnvContainHosts(envName string) bool {

	if len(RGetHostsInEnv(envName)) > 0 {
		return true
	} else {
		return false
	}
}

func DoLock(c *CacheDataType, res *WebResponseType) bool {

	c.State = C_STATE_LOCKED

	if c.Type == C_TYPE_ENV {

		c.Parent = "n/a"
		res.Messages = append(res.Messages, OK_EnvLocked)

		if RSetLockData(c) { // OK
			c.HttpErr = http.StatusOK
			return true

		} else { // ERROR
			res.Messages = append(res.Messages, ERR_EnvLockFail)
			c.HttpErr = http.StatusForbidden
			return false

		}

	} else if c.Type == C_TYPE_HOST {

		c.Parent = GetEnvFromHost(c.Name)

		if IsEnvLocked(c.Parent) { // parent env has been locked
			res.Messages = append(res.Messages, ERR_ParentEnvLockFail)
			c.HttpErr = http.StatusForbidden
			return false
		}

		res.Messages = append(res.Messages, OK_HostLocked)

		if RSetLockData(c) { // OK
			c.HttpErr = http.StatusOK
			return true

		} else { // ERROR
			res.Messages = append(res.Messages, ERR_HostLockFail)
			c.HttpErr = http.StatusForbidden
			return false
		}

	} else { // illegal type

		res.Messages = append(res.Messages, ERR_WrongTypeSpecified)
		c.HttpErr = http.StatusInternalServerError
		return false
	}
}
