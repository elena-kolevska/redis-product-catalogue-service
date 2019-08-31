package main

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"strconv"
	"strings"
)

func getCategoryNameById(id int) (string, error) {
	categoryName, err := redis.String(redisConn.Do("HGET", config.KeyCategories, id))
	if err != nil {
		return "", err
	}
	return categoryName, nil
}

func getImageNameById(id int) string {
	return fmt.Sprintf(config.KeyImage, strconv.Itoa(id))
}

func getImageUrlById(id int) string {
	return fmt.Sprintf("/images/%v", strconv.Itoa(id))
}

func getProductImagesKeyName(id int) string {
	return fmt.Sprintf(config.KeyProductImages, strconv.Itoa(id))
}

func populateProductFromHash(values []interface{}) (Product, error) {
	var product Product
	err := redis.ScanStruct(values, &product)
	if err != nil {
		return product, err
	}

	//////////////////////////////////////////
	// Get the product category and attach it to the product struct
	//////////////////////////////////////////
	product.setCategory()

	return product, nil
}

func getCategoriesMap() map[int]Category {
	categories := make(map[int]Category, 0)
	values, _ := redis.StringMap(redisConn.Do("HGETALL", config.KeyCategories))

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

func normaliseSearchString(s string) string {
	return strings.Replace(strings.ToLower(s), "::", " ", -1)
}