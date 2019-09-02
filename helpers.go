package main

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"strconv"
	"strings"
)

func getCategoryNameById(id int, redisConn redis.Conn) (string, error) {
	categoryName, err := redis.String(redisConn.Do("HGET", config.KeyCategories, id))
	if err != nil {
		return "", err
	}
	return categoryName, nil
}

func getProductNameById(id int) string {
	return fmt.Sprintf(config.KeyProduct, id)
}

func getNextImageId(redisConn redis.Conn) int {
	id, _ := redis.Int(redisConn.Do("INCR", config.KeyImageCounter))
	return id
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
func getProductsInCategoryKeyName(categoryId int) string {
	return fmt.Sprintf(config.KeyProductsInCategory, categoryId)
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

func getProductImagesFromHash(values map[string]string) []Image {
	images := make([]Image, 0)

	for imageId, imageUrl := range values {
		imageId, _ := strconv.Atoi(imageId)
		image := Image{
			Id:   imageId,
			Url: config.BaseUri + imageUrl,
		}
		images = append(images, image)
	}

	return images
}

func normaliseSearchString(s string) string {
	return strings.Replace(strings.ToLower(s), "::", " ", -1)
}

func getHashAsStringMap (keyName string, redisConn redis.Conn) (map[string]string, error) {
	return redis.StringMap(redisConn.Do("HGETALL", keyName))
}