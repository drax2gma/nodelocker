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

	if err != nil {
		return false
	}

	errExp := RConn.Expire(key, lastDay).Err()
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

func RLog(msg string) bool {

	err := RConn.RPush("log", msg).Err()
	if err != nil {
		log.Fatal(err.Error())
	}

	// keeping log length in a sane interval (1000 lines)
	RConn.LTrim("log", 0, -1000)

	return true
}

// chkType: C_USER_Exists or C_USER_Valid
func RCheckUser(chkType string, userName string, userToken string) bool {

	redisPwd, err := RConn.HGet("users", userName).Result()
	if err != nil {

		return false

	} else {

		if chkType == C_USER_Exists {

			return true

		} else if chkType == C_USER_Valid {

			hashedPwd := CryptString(userToken)
			return hashedPwd == redisPwd
		}
	}

	return false
}
