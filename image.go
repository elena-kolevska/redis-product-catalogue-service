package main

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"strconv"
)

type Image struct {
	Id        int    `json:"id"`
	ProductId int    `json:"product_id,omitempty"`
	Url       string `json:"url"`
}

func (image *Image) setId(redisConn redis.Conn) {
	id, _ := redis.Int(redisConn.Do("INCR", config.KeyImageCounter))
	image.Id = id
}

func (image *Image) setUrl() {
	image.Url = config.BaseUri + fmt.Sprintf("/images/%v", strconv.Itoa(image.Id))
}

func (image *Image) delete(redisConn redis.Conn) {
	// Start a transaction and send all commands in a pipeline
	_, _ = redisConn.Do("MULTI")
	_, _ = redisConn.Do("SREM", getProductImagesKeyName(image.ProductId), image.Id)
	_, _ = redisConn.Do("HDEL", config.KeyImages, image.Id)
	_, _ = redisConn.Do("DEL", getImageNameById(image.Id))
	_, _ = redisConn.Do("EXEC")
}

func getImageDataById(id int, redisConn redis.Conn) ([]byte, error) {
	return redis.Bytes(redisConn.Do("GET", getImageNameById(id)))
}

func saveNewImage(productId int, data []byte, redisConn redis.Conn) (Image, error) {
	image := Image {
		ProductId: productId,
	}

	image.setId(redisConn)
	image.setUrl()

	// Create image key name
	keyName := getImageNameById(image.Id)

	_, err := redisConn.Do("SET", keyName, data)
	if err != nil {
		return Image{}, err
	}

	// Save image to "all images" hash
	_, err = redisConn.Do("HSET", config.KeyImages, image.Id, productId)
	if err != nil {
		return Image{}, err
	}


	// Add image to product's images hash
	productImagesKeyName := getProductImagesKeyName(productId)
	_, err = redisConn.Do("SADD", productImagesKeyName, image.Id)
	if err != nil {
		return image, err
	}

	return image, nil
}

func getImageNameById(id int) string {
	return fmt.Sprintf(config.KeyImage, strconv.Itoa(id))
}