package x

import (
	"fmt"
	"log"
	"net/http"
	"sort"
	"time"

	"github.com/go-redis/redis"
)

var (
	RConn *redis.Client // global Redis connection

)

func RGetSingle(key string, field string) any {

	value, err := RConn.HGet(key, field).Result()
	if err != nil || value == "" {

		// In this function we use 'nil' as false return value
		// if something gone wrong
		return nil
	}

	return value
}

func RSetSingle(key string, field string, value any, lastDay time.Duration) bool {

	err := RConn.HSet(key, field, value).Err()
	return err == nil
}

// Usually key := 'entityType:entityName'
func RSetExpire(key string, expire time.Duration) bool {

	errExp := RConn.Expire(key, expire).Err()
	return errExp == nil
}

// Usually key := 'entityType:entityName'
func RLockGetter(key string) *LockData {

	c := new(LockData)
	var resultsMap map[string]string

	result, err := RConn.HMGet(key, "parent", "state", "user", "lastday").Result()
	if err != nil || result[0] == nil { // no record found
		c.HttpErr = http.StatusNoContent
		return c
	}

	fields := []string{"parent", "state", "user", "lastday"}
	resultsMap = make(map[string]string)

	for i, field := range fields {
		resultsMap[field] = result[i].(string)
	}

	c.Parent = resultsMap["parent"]
	c.State = resultsMap["state"]
	c.User = resultsMap["user"]
	c.LastDay = resultsMap["lastday"]
	c.HttpErr = http.StatusOK

	return c
}

// Do not forget to fill x.LockData before function call!
//
// Returns `true` on successful run.
func RLockSetter(c *LockData) bool {

	err := RConn.HMSet(c.Type+":"+c.Name, map[string]any{
		"state":   c.State,
		"parent":  c.Parent,
		"user":    c.User,
		"lastday": c.LastDay,
	}).Err()

	return err == nil
}

func REntityDelete(enType string, enName string) bool {

	err := RConn.HDel(enType, enName).Err()
	if err != nil {
		log.Fatal(err.Error())
	}

	// Check if entity has gone
	if RGetSingle(enType, enName) == nil {
		return true
	} else {
		return false
	}
}

func RGetHostsInEnv(envName string) []string {

	// Create a cursor to iterate over hash sets
	var cursor uint64 = 0
	var resultList []string

	fmt.Printf("Checking hosts in '%s' environment...", envName)

	for {
		// Use SCAN command to get keys matching the pattern
		keys, nextCursor, err := RConn.Scan(cursor, C_TYPE_HOST+":*", 0).Result()
		if err != nil {
			log.Fatal(err.Error())
		}

		// Loop through the keys and check if 'env' field contains envName
		for _, key := range keys {
			// Use HGET command to get the value of the 'env' field
			parent, err := RConn.HGet(key, C_PARENT).Result()
			if err != nil {
				log.Fatal(err.Error())
			}

			if parent == envName {
				resultList = append(resultList, key)
				fmt.Printf("%s ", key)
			}
		}
		fmt.Println()

		// If nextCursor is 0, it means iteration is complete
		if nextCursor == 0 {
			break
		}

		// Update the cursor for the next iteration
		cursor = nextCursor
	}

	return resultList

}

// matchPattern should be C_TYPE_ENV or C_TYPE_HOST
func RScanKeys(matchPattern string) []string {

	var cursor uint64
	keys := make([]string, 0)

	for {
		var (
			result []string
			err    error
		)
		result, cursor, err = RConn.Scan(cursor, matchPattern+":*", 10).Result()
		if err != nil {
			log.Fatal(err.Error())
		}
		keys = append(keys, result...)
		if cursor == 0 {
			break
		}
	}
	sort.Strings(keys)
	return keys
}

func RFillJsonStats(r *Stats) {

	envs := RScanKeys(C_TYPE_ENV)
	envPrefixLen := len(C_TYPE_ENV) + 1
	hostPrefixLen := len(C_TYPE_HOST) + 1

	for _, key := range envs { // environment data

		var envName string
		var lockingUser string
		var lastDay string

		result, err := RConn.HGetAll(key).Result()
		if err != nil {
			fmt.Printf("Error fetching data for key %s: %s\n", key, err)
			continue
		}

		for field, value := range result {

			switch {
			case field == "state" && value == C_STATE_VALID:
				r.ValidEnvs = append(r.ValidEnvs, key[envPrefixLen:])
			case field == "state" && value == C_STATE_LOCKED:
				envName = key[envPrefixLen:]
				if len(lockingUser) > 0 {
					h := envName + " (👤" + lockingUser + "   📅" + lastDay + ")"
					r.LockedEnvs = append(r.LockedEnvs, h)
				}
			case field == "state" && value == C_STATE_MAINTENANCE:
				r.MaintEnvs = append(r.MaintEnvs, key[envPrefixLen:])
			case field == "state" && value == C_STATE_TERMINATED:
				r.TermdEnvs = append(r.TermdEnvs, key[envPrefixLen:])
			case field == "user":
				lockingUser = value
			case field == "lastday":
				lastDay = value
			}
		}
	}

	hosts := RScanKeys(C_TYPE_HOST)

	for _, key := range hosts { // host data

		var hostName string
		var lockingUser string
		var lastDay string

		result, err := RConn.HGetAll(key).Result()
		if err != nil {
			fmt.Printf("Error fetching data for key %s: %s\n", key, err)
			continue
		}

		for field, value := range result {

			switch {
			case field == "state" && value == C_STATE_LOCKED:
				hostName = key[hostPrefixLen:]
			case field == "user":
				lockingUser = value
			case field == "lastday":
				lastDay = value
			}
		}

		h := hostName + " (👤" + lockingUser + "   📅" + lastDay + ")"
		r.LockedHosts = append(r.LockedHosts, h)
	}
}
