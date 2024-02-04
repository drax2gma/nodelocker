package shared

import (
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis"
)

var (
	RConn *redis.Client
)

func RedisGet(key string, field string) (any, string) {

	value, err := RConn.HGet(key, field).Result()
	if err != nil || value == "" {
		return "No value", "NIL"
	}

	return value, "OK"
}

func RedisSet(key string, field string, value any, expire time.Duration) (any, string) {

	err := RConn.HSet(key, field, value).Err()

	// _, err := RConn.Do("hset", field, map[string]interface{}{
	// 	"title":  title,
	// 	"link":   link,
	// 	"poster": user,
	// 	"time":   now,
	// 	"votes":  1,
	// }).Result()

	if err != nil {
		return "SET failed", "ERR"
	}

	errExp := RConn.Expire(key, expire).Err()
	if errExp != nil {
		return "EXPIRE set failed", "ERR"
	}

	return fmt.Sprintf("%s == %s", key, value), "OK"

}

func RedisLog(msg string) {

	var maxLogLines int64 = 1000

	err := RConn.RPush("log", msg).Err()
	if err != nil {
		log.Fatal(err.Error())
	}

	// keeping log length in a sane interval
	RConn.LTrim("log", 0, 0-maxLogLines)

}

func RedisValidUser(user string, token string) bool {

	redisPwd, err := RConn.HGet("users", user).Result()
	if err != nil || redisPwd == "" {
		return false
	}

	hashedPwd := CryptString(token)

	return hashedPwd == redisPwd

}
