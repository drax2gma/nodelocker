package x

import (
	"crypto/sha1"
	"fmt"
	"log"
	"regexp"
	"time"
)

type JsonDataType struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Type    string `json:"type"`
	Name    string `json:"name"`
	State   string `json:"state"`
	LastDay string `json:"lastday"`
	User    string `json:"user"`
	Token   string `json:"token"`
}

const (
	C_CacheData string = "(cached)"
)

var CacheData JsonDataType // temp manipulation of data fields

// Crypt a plain string into sha1 hash string
func CryptString(plain string) string {

	h := sha1.New()
	h.Write([]byte(plain))
	sha1Hash := h.Sum(nil)
	hexString := fmt.Sprintf("%x", sha1Hash)
	return hexString
}

func ResetWebResponse(res *JsonDataType) {
	res.Type = ""
	res.Name = ""
	res.State = ""
	res.User = ""
	res.Success = false
	res.Message = "ERR: Unexpected error."
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
