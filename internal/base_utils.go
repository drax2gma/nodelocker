package x

import (
	"crypto/sha1"
	"fmt"
	"log"
	"regexp"
	"slices"
	"time"
)

type CacheDataType struct {
	Type    string `json:"type"`
	Name    string `json:"name"`
	State   string `json:"state"`
	LastDay string `json:"lastday"`
	User    string `json:"user"`
	Token   string `json:"token"`
	HttpErr int    `json:"httperr"`
}

type WebResponseType struct {
	Success  bool     `json:"success"`
	Messages []string `json:"messages"`
	Type     string   `json:"type"`
	Name     string   `json:"name"`
	State    string   `json:"state"`
	LastDay  string   `json:"lastday"`
	User     string   `json:"user"`
}

const (
	C_ADMIN     string = "admin"
	C_ENV_LIST  string = "envlist"
	C_TYPE_ENV  string = "env"
	C_TYPE_HOST string = "host"

	C_SUCCESS string = "✅"
	C_FAILED  string = "❌"
	C_STARTED string = "🌐"

	C_LOCKED      string = "locked"
	C_TERMINATED  string = "termnd"
	C_MAINTENANCE string = "maint"

	C_RespHeader string = "application/json"
	C_Secret     string = "XXXXXXX"
)

func NewWebResponse() WebResponseType {

	var r WebResponseType

	r.Type = ""
	r.Name = ""
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
		fmt.Println("regex nomatch")
		return false // Input doesn't match the pattern
	}

	fmt.Println("regex MATCH")

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

func IsValidEntityType(t string) bool {

	validTypes := []string{C_TYPE_ENV, C_TYPE_HOST}
	return slices.Contains(validTypes, t)

}

func IsExistingUser(userName string) bool {

	fmt.Printf("Checking user '%s'... ", userName)
	_, err := RConn.HGet("user", userName).Result()

	if err == nil {
		fmt.Printf("found.\n")
		return true
	} else {
		fmt.Printf("not found.\n")
		return false
	}
}

func IsValidUser(userName string, userToken string) bool {

	fmt.Printf("Validating user '%s'... ", userName)
	redisPwd, err := RConn.HGet("user", userName).Result()

	if err == nil && len(redisPwd) > 0 {
		// found user & password

		if redisPwd == CryptString(userToken) {
			fmt.Printf("user is valid.\n")
			return true
		} else {
			fmt.Printf("password mismatch.\n")
			return false
		}
	} else {
		fmt.Printf("user not found.\n")
		return false
	}
}

// acquire the env from the hostname automagically
func GetEnvFromHost(hostName string) string {

	separators := []string{"-", "_", "/", ".", ",", ":"}

	for i, char := range hostName {
		for _, separator := range separators {
			if string(char) == separator {
				return hostName[:i]
			}
		}
	}
	return hostName
}

func IsEnvLocked(envName string) bool {

	c := NewCacheData()
	c.Type = C_TYPE_ENV
	c.Name = envName
	return RGetLockData(&c)
}

func CreateEnv(envName string) bool {

	return RSetAddMember(C_ENV_LIST, envName)
}

func MaintenanceEnv(envName string) bool {

	c := NewCacheData()
	c.Type = C_TYPE_ENV
	c.Name = envName
	c.State = C_MAINTENANCE
	return RSetLockData(&c)
}

func TerminateEnv(envName string) bool {

	c := NewCacheData()
	c.Type = C_TYPE_ENV
	c.Name = envName
	c.State = C_TERMINATED
	p1 := RGetLockData(&c)
	p2 := RSetRemoveMember(C_TYPE_ENV, envName)

	return p1 && p2
}

func UnlockEnv(envName string) bool {

	return REntityDelete(C_TYPE_ENV, envName)

}

func UnlockHost(hostName string) bool {

	return REntityDelete(C_TYPE_HOST, hostName)

}
