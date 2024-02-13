package x

import (
	"crypto/sha1"
	"fmt"
	"log"
	"net/http"
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
	C_ADMIN          string = "admin"
	C_UseCacheData   string = "(cached)"
	C_USER_Valid     string = "uv"
	C_USER_Invalid   string = "ui"
	C_USER_Exists    string = "ue"
	C_USER_NotExists string = "un"
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
	c.HttpErr = http.StatusInternalServerError

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

	validTypes := []string{"env", "host"}
	return slices.Contains(validTypes, t)

}
