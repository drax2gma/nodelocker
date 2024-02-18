package x

import (
	"fmt"
	"log"
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

// Do not forget to fill x.CacheData before function call!
func RGetLockData(c *CacheDataType) bool {

	var resultsMap map[string]string

	result, err := RConn.HMGet(c.Type+":"+c.Name, "state", "user", "lastday").Result()
	if err != nil || result[0] == nil {
		return false
	}

	fields := []string{"state", "user", "lastday"}
	resultsMap = make(map[string]string)

	for i, field := range fields {
		resultsMap[field] = result[i].(string)
	}

	c.State = resultsMap["state"]
	c.Parent = resultsMap["parent"]
	c.User = resultsMap["user"]
	c.LastDay = resultsMap["lastday"]

	return true
}

// Do not forget to fill x.CacheData before function call!
func RSetLockData(c *CacheDataType) bool {

	err := RConn.HMSet(c.Type+":"+c.Name, map[string]interface{}{
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
		fmt.Println(err.Error())
	}

	// Check if entity has gone
	if RGetSingle(enType, enName) == nil {
		return true
	} else {
		return false
	}
}

// func RSetAddMember(setName string, member string) bool {

// 	err := RConn.SAdd(setName, member).Err()
// 	if err != nil {
// 		fmt.Println(err.Error())
// 	}
// 	return true
// }

// func RSetRemoveMember(setName string, member string) bool {

// 	err := RConn.SRem(setName, member).Err()
// 	if err != nil {
// 		fmt.Println(err.Error())
// 	}
// 	return true
// }

// func RIsMemberOfSet(setName string, member string) bool {

// 	m, err := RConn.SIsMember(setName, member).Result()
// 	if err != nil {
// 		fmt.Println(err.Error())
// 	}
// 	return m
// }

func RGetHostsInEnv(envName string) []string {

	// Create a cursor to iterate over hash sets
	var cursor uint64 = 0
	var resultList []string

	fmt.Printf("Checking hosts in %s env...", envName)

	for {
		// Use SCAN command to get keys matching the pattern
		keys, nextCursor, err := RConn.Scan(cursor, C_TYPE_HOST+":*", 0).Result()
		if err != nil {
			fmt.Println(err.Error())
		}

		// Loop through the keys and check if 'env' field contains envName
		for _, key := range keys {
			// Use HGET command to get the value of the 'env' field
			parent, err := RConn.HGet(key, C_PARENT).Result()
			if err != nil {
				fmt.Println(err.Error())
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

func RFillJsonStats(r *StatsType) {

	envs := RScanKeys(C_TYPE_ENV)
	envPrefixLen := len(C_TYPE_ENV) + 1
	hostPrefixLen := len(C_TYPE_HOST) + 1

	for _, key := range envs {
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
				r.LockedEnvs = append(r.LockedEnvs, key[envPrefixLen:])
			case field == "state" && value == C_STATE_MAINTENANCE:
				r.MaintEnvs = append(r.MaintEnvs, key[envPrefixLen:])
			case field == "state" && value == C_STATE_TERMINATED:
				r.TermdEnvs = append(r.TermdEnvs, key[envPrefixLen:])
			}
		}
	}

	hosts := RScanKeys(C_TYPE_HOST)

	for _, key := range hosts {
		result, err := RConn.HGetAll(key).Result()
		if err != nil {
			fmt.Printf("Error fetching data for key %s: %s\n", key, err)
			continue
		}

		for field, value := range result {

			switch {
			case field == "state" && value == C_STATE_LOCKED:
				r.LockedHosts = append(r.LockedHosts, key[hostPrefixLen:])
			}
		}
	}
}
