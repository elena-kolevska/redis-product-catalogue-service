package main

import (
	"github.com/gomodule/redigo/redis"
	"strconv"
)

type Category struct {
	Id   int    `redis:"id" json:"id"`
	Name string `redis:"name" json:"name"`
}

func getCategoriesMap(redisConn redis.Conn) map[int]Category {
	categories := make(map[int]Category, 0)
	values, _ := getHashAsStringMap(config.KeyCategories, redisConn)

	for categoryId, categoryName := range values {
		categoryId, _ := strconv.Atoi(categoryId)
		category := Category{
			Id:   categoryId,
			Name: categoryName,
		}
		categories[categoryId] = category
	}

	return categories
}


func getCategoryNameById(id int, redisConn redis.Conn) (string, error) {
	categoryName, err := redis.String(redisConn.Do("HGET", config.KeyCategories, id))
	if err != nil {
		return "", err
	}
	return categoryName, nil
}