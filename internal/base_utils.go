package shared

import (
	"crypto/sha1"
	"fmt"
	"log"
	"time"
)

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
