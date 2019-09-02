package main

import (
	"github.com/gomodule/redigo/redis"
	"strings"
)

func normaliseSearchString(s string) string {
	return strings.Replace(strings.ToLower(s), "::", " ", -1)
}

func getHashAsStringMap (keyName string, redisConn redis.Conn) (map[string]string, error) {
	return redis.StringMap(redisConn.Do("HGETALL", keyName))
}