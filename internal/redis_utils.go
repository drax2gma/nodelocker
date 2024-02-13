package x

import (
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis"
)

var (
	RConn *redis.Client // global Redis connection

)

// Usually key := 'entityType:entityName'
func RGetSingle(key string, field string) any {

	value, err := RConn.HGet(key, field).Result()
	if err != nil || value == "" {

		// In this function we use 'nil' as false return value
		// if something gone wrong
		return nil
	}

	return value
}

// Usually key := 'entityType:entityName'
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
	c.User = resultsMap["user"]
	c.LastDay = resultsMap["lastday"]

	return true
}

// Do not forget to fill x.CacheData before function call!
func RSetLockData(c *CacheDataType) bool {

	err := RConn.HMSet(c.Type+":"+c.Name, map[string]interface{}{
		"state":   c.State,
		"user":    c.User,
		"lastday": c.LastDay,
	}).Err()

	return err == nil
}

func REntityDelete(enType string, enName string) bool {

	err := RConn.Del(enType + ":" + enName).Err()
	if err != nil {
		log.Fatal(err.Error())
	}

	return true
}

// chkType: C_USER_Exists or C_USER_Valid
// returns: C_USER_Exists, C_USER_NotExists, C_USER_Valid, C_USER_Invalid
func RCheckUser(chkType string, userName string, userToken string) string {

	fmt.Println("Checking user: " + userName)
	redisPwd, err := RConn.HGet("user", userName).Result()

	if err == nil { // found something

		if chkType == C_USER_Exists {

			fmt.Println(userName + " user found.")
			return C_USER_Exists

		} else if chkType == C_USER_Valid {

			if redisPwd == CryptString(userToken) {

				fmt.Println(userName + " user is valid.")
				return C_USER_Valid

			} else {

				fmt.Println(userName + " user is invalid.")
				return C_USER_Invalid
			}

		} else {

			fmt.Println(userName + " user not found.")
			return C_USER_NotExists
		}

	}

	// nothing in db
	fmt.Println(userName + " user not exists (redis error).")
	return C_USER_NotExists
}
