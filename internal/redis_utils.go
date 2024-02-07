package util

import (
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis"
)

const (
	C_SUBJECT = "subject"
	C_STATE   = "state"
	C_OWNER   = "owner"
	C_EXPIRE  = "expire"
)

var (
	RConn             *redis.Client // global Redis connection
	RedisLastErrorMsg string        // Storage for the last Redis error for debug
)

func RedisGet(key string, field string) any {
	// In this function we use 'nil' as false return value
	// if something gone wrong

	value, err := RConn.HGet(key, field).Result()
	if err != nil || value == "" {
		RedisLastErrorMsg = "ERR: No value"
		return nil
	}

	return value
}

func RedisSet(key string, field string, value any, expire time.Duration) bool {

	err := RConn.HSet(key, field, value).Err()

	if err != nil {
		RedisLastErrorMsg = "ERR: SET failed"
		return false
	}

	errExp := RConn.Expire(key, expire).Err()
	if errExp != nil {
		RedisLastErrorMsg = "ERR: EXPIRE set failed"
		return false
	}

	return true
}

func RedisGetLockData() bool {

	var resultsMap map[string]string

	result, err := RConn.HMGet(LockDataSUBJECT, C_STATE, C_OWNER, C_EXPIRE).Result()
	if err != nil || result[0] == nil {
		RedisLastErrorMsg = "ERR: HMGet failed, empty result."
		return false
	}

	fmt.Println(result, err)

	fields := []string{C_STATE, C_OWNER, C_EXPIRE}
	resultsMap = make(map[string]string)

	for i, field := range fields {
		resultsMap[field] = result[i].(string)
	}

	LockDataSTATE = resultsMap[C_STATE]
	LockDataOWNER = resultsMap[C_OWNER]
	LockDataEXPIRE = resultsMap[C_EXPIRE]

	return true
}

func RedisSetLockData() bool {

	err := RConn.HMSet(LockDataSUBJECT, map[string]interface{}{
		C_STATE:  LockDataSTATE,
		C_OWNER:  LockDataOWNER,
		C_EXPIRE: LockDataEXPIRE,
	}).Err()

	if err != nil {
		RedisLastErrorMsg = "ERR: HMSet failed"
		return false
	}

	return true
}

func RedisDelete() bool {

	err := RConn.Del(LockDataSUBJECT).Err()
	if err != nil {
		log.Fatal(err.Error())
	}

	return true
}

func RedisLog(msg string) bool {

	var maxLogLines int64 = 1000

	err := RConn.RPush("log", msg).Err()
	if err != nil {
		log.Fatal(err.Error())
	}

	// keeping log length in a sane interval
	RConn.LTrim("log", 0, 0-maxLogLines)

	return true
}

func RedisValidUser(user string, token string) bool {

	redisPwd, err := RConn.HGet("users", user).Result()
	if err != nil || redisPwd == "" {
		RedisLastErrorMsg = "ERR: Illegal user"
		return false
	}

	hashedPwd := CryptString(token)

	return hashedPwd == redisPwd
}
