package x

import (
	"log"
	"time"

	"github.com/go-redis/redis"
)

var (
	RConn *redis.Client // global Redis connection

)

func RGetSingle(key string, field string) any {

	if key == C_UseCacheData {
		key = CacheData.Type + ":" + CacheData.Name
	}

	value, err := RConn.HGet(key, field).Result()
	if err != nil || value == "" {

		// In this function we use 'nil' as false return value
		// if something gone wrong
		return nil
	}

	return value
}

func RSetSingle(key string, field string, value any, lastDay time.Duration) bool {

	if key == C_UseCacheData {
		key = CacheData.Type + ":" + CacheData.Name
	}

	err := RConn.HSet(key, field, value).Err()
	return err == nil
}

func RSetExpire(key string, expire time.Duration) bool {

	errExp := RConn.Expire(key, expire).Err()
	return errExp == nil
}

// Do not forget to fill x.CacheData before function call!
func RGetLockData() bool {

	var resultsMap map[string]string

	result, err := RConn.HMGet(CacheData.Type+":"+CacheData.Name, "state", "user", "lastday").Result()
	if err != nil || result[0] == nil {
		return false
	}

	fields := []string{"state", "user", "lastday"}
	resultsMap = make(map[string]string)

	for i, field := range fields {
		resultsMap[field] = result[i].(string)
	}

	CacheData.State = resultsMap["state"]
	CacheData.User = resultsMap["user"]
	CacheData.LastDay = resultsMap["lastday"]

	return true
}

// Do not forget to fill x.CacheData before function call!
func RSetLockData() bool {

	err := RConn.HMSet(CacheData.Type+":"+CacheData.Name, map[string]interface{}{
		"state":   CacheData.State,
		"user":    CacheData.User,
		"lastday": CacheData.LastDay,
	}).Err()

	return err == nil
}

func REntityDelete() bool {

	err := RConn.Del(CacheData.Type + ":" + CacheData.Name).Err()
	if err != nil {
		log.Fatal(err.Error())
	}

	return true
}

// chkType: C_USER_Exists or C_USER_Valid
// returns: C_USER_Exists, C_USER_NotExists, C_USER_Valid, C_USER_Invalid
func RCheckUser(chkType string, userName string, userToken string) string {

	CLog("Checking user: " + userName)
	redisPwd, err := RConn.HGet("user", userName).Result()

	if err == nil { // found something

		if chkType == C_USER_Exists {

			CLog(userName + " user found.")
			return C_USER_Exists

		} else if chkType == C_USER_Valid {

			if redisPwd == CryptString(userToken) {

				CLog(userName + " user is valid.")
				return C_USER_Valid

			} else {

				CLog(userName + " user is invalid.")
				return C_USER_Invalid
			}

		} else {

			CLog(userName + " user not found.")
			return C_USER_NotExists
		}

	}

	// nothing in db
	CLog(userName + " user not exists (redis error).")
	return C_USER_NotExists
}
