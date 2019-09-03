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

func (image *Image) delete(redisConn redis.Conn) error {

	// Start a transaction and send all commands in a pipeline
	_, err := redisConn.Do("MULTI")
	if err != nil {
		return err
	}

	_ = redisConn.Send("SREM", getProductImagesKeyName(image.ProductId), image.Id)
	_ = redisConn.Send("HDEL", config.KeyImages, image.Id)
	_ = redisConn.Send("DEL", getImageNameById(image.Id))

	_, err = redisConn.Do("EXEC")
	if err != nil {
		return err
	}

	return nil
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

	// Start a transaction and send all commands in a pipeline
	_, err := redisConn.Do("MULTI")
	if err != nil {
		return Image{}, err
	}

	_ = redisConn.Send("SET", keyName, data)

	// Save image to "all images" hash
	_ = redisConn.Send("HSET", config.KeyImages, image.Id, productId)


	// Add image to product's images hash
	productImagesKeyName := getProductImagesKeyName(productId)
	_ = redisConn.Send("SADD", productImagesKeyName, image.Id)

	_, err = redisConn.Do("EXEC")
	if err != nil {
		return Image{}, err
	}

	return image, nil
}

func getImageNameById(id int) string {
	return fmt.Sprintf(config.KeyImage, strconv.Itoa(id))
}