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