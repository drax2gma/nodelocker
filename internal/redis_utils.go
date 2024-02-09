package x

import (
	"log"
	"time"

	"github.com/go-redis/redis"
)

var (
	RConn         *redis.Client // global Redis connection
	RLastErrorMsg string        // Storage for the last Redis error for debug
)

func RGetSingle(key string, field string) any {
	// In this function we use 'nil' as false return value
	// if something gone wrong

	value, err := RConn.HGet(key, field).Result()
	if err != nil || value == "" {
		RLastErrorMsg = "ERR: No value"
		return nil
	}

	return value
}

func RSetSingle(key string, field string, value any, lastDay time.Duration) bool {

	if key == C_CacheData {
		key = CacheData.Type + ":" + CacheData.Name
	}

	err := RConn.HSet(key, field, value).Err()

	if err != nil {
		RLastErrorMsg = "ERR: SET failed"
		return false
	}

	errExp := RConn.Expire(key, lastDay).Err()
	if errExp != nil {
		RLastErrorMsg = "ERR: EXPIRE set failed"
		return false
	}

	return true
}

// Do not forget to fill util.CacheData before function call!
func RGetLockData() bool {

	var resultsMap map[string]string

	result, err := RConn.HMGet(CacheData.Type+":"+CacheData.Name, "state", "user", "lastday").Result()
	if err != nil || result[0] == nil {
		RLastErrorMsg = "ERR: HMGet failed, empty result."
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

// Do not forget to fill util.CacheData before function call!
func RSetLockData() bool {

	err := RConn.HMSet(CacheData.Type+":"+CacheData.Name, map[string]interface{}{
		"state":   CacheData.State,
		"user":    CacheData.User,
		"lastday": CacheData.LastDay,
	}).Err()

	if err != nil {
		RLastErrorMsg = "ERR: HMSet failed"
		return false
	}

	return true
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

func RValidUser(user string, token string) bool {

	redisPwd, err := RConn.HGet("users", user).Result()
	if err != nil || redisPwd == "" {
		RLastErrorMsg = "ERR: Illegal user"
		return false
	}

	hashedPwd := CryptString(token)

	return hashedPwd == redisPwd
}
