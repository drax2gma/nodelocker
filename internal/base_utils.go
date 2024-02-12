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
}

type WebResponseDataType struct {
	Success  bool     `json:"success"`
	Messages []string `json:"messages"`
	Type     string   `json:"type"`
	Name     string   `json:"name"`
	State    string   `json:"state"`
	LastDay  string   `json:"lastday"`
	User     string   `json:"user"`
}

type StatusRespType struct {
	Success  bool     `json:"success"`
	Messages []string `json:"messages"`
}

const (
	C_UseCacheData   string = "(cached)"
	C_USER_Valid     string = "uv"
	C_USER_Invalid   string = "ui"
	C_USER_Exists    string = "ue"
	C_USER_NotExists string = "un"
)

var CacheData CacheDataType // temp manipulation of data fields

func ResetWebResponse(r *WebResponseDataType) {
	r.Type = ""
	r.Name = ""
	r.State = ""
	r.LastDay = ""
	r.User = ""
	r.Success = false
	r.Messages = []string{}
}

func ResetCacheData(c *CacheDataType) {
	c.Type = ""
	c.Name = ""
	c.State = ""
	c.LastDay = ""
	c.User = ""
	c.Token = ""
}

func ResetStatusResponse(r *StatusRespType) {
	r.Success = false
	r.Messages = []string{}
}

// Crypt a plain string into sha1 hash string
func CryptString(plain string) string {

	h := sha1.New()
	h.Write([]byte(plain))
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

func CLog(msg string) {

	fmt.Println(msg)

}
