package main

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"github.com/labstack/gommon/log"
	"strconv"
)

//////////////////////
// PRODUCT MODEL
//////////////////////
type Product struct {
	Id             int      `redis:"id" json:"id"`
	Name           string   `redis:"name" json:"name"`
	Description    string   `redis:"description" json:"description"`
	Vendor         string   `redis:"vendor" json:"vendor"`
	Price          float32  `redis:"price" json:"price"`
	Currency       string   `redis:"currency" json:"currency"`
	MainCategoryId int      `redis:"main_category_id" json:"main_category_id,omitempty"`
	MainCategory   Category `redis:"-" json:"main_category"`
	Images         []Image  `redis:"-" json:"images" `
}

func (product *Product) setId() {
	id, _ := redis.Int(redisConn.Do("INCR", config.KeyProductCounter))
	product.Id = id
}
func (product *Product) getKeyName() string {
	return fmt.Sprintf(config.KeyProduct, product.Id)
}
func (product *Product) getLexName() string {
	return fmt.Sprintf("%s::%v", product.getNormalisedName(), product.Id)
}
func (product *Product) getNormalisedName() string {
	return normaliseSearchString(product.Name)
}
func (product *Product) setCategory() {
	categoryName, _ := redis.String(redisConn.Do("HGET", config.KeyCategories, product.MainCategoryId))
	category := Category{
		Id:   product.MainCategoryId,
		Name: categoryName,
	}
	product.setCategoryFromStruct(category)
}

func (product *Product) setCategoryFromStruct(category Category) {
	product.MainCategory = category
	product.MainCategoryId = 0 //We don't want to show this field directly on the product object, but as a part of its category
}

func (product *Product) delete(redisConn redis.Conn) error {

	productValues, err := redis.Values(redisConn.Do("HGETALL", product.getKeyName()))
	if err != nil {
		log.Error(err)
		return err
	}
	if len(productValues) == 0 {
		log.Error(err)
		return &notFoundError
	}
	err = redis.ScanStruct(productValues, product)
	if err != nil {
		log.Error(err)
		return err
	}

	// Delete all product images
	productImagesKeyName := getProductImagesKeyName(product.Id)
	imageValues, _ := redis.StringMap(redisConn.Do("HGETALL", productImagesKeyName))

	// Start a transaction and send all commands in a pipeline
	_,_ = redisConn.Do("MULTI")

	for imageId, _ := range imageValues {
		imageId, _ := strconv.Atoi(imageId)
		_ = redisConn.Send("DEL", getImageNameById(imageId))
		// Delete from "all images" hash
		_ = redisConn.Send("HDEL", config.KeyImages, imageId)
	}

	// Delete the product images hash
	_ = redisConn.Send("DEL", productImagesKeyName)


	// Delete from the all_products and "products_by_cat" hashes
	_ = redisConn.Send("ZREM", config.KeyAllProducts, product.getLexName())
	_ = redisConn.Send("ZREM", getProductsInCategoryKeyName(product.MainCategoryId), product.getLexName())

	// Delete the image key
	_ = redisConn.Send("DEL", product.getKeyName())

	// Execute transaction
	_, err = redisConn.Do("EXEC")
	if err != nil {
		return err
	}

	return nil
}

func getProductById(id int, redisConn redis.Conn) (Product, error) {

	//////////////////////////////////////////
	// Fetch the details of a specific product.
	//////////////////////////////////////////
	values, err := redis.Values(redisConn.Do("HGETALL", getProductNameById(id)))
	if err != nil {
		return Product{}, err
	}
	// If no product is found for the given id, return a 404
	if len(values) == 0 {
		return Product{}, &notFoundError
	}

	//////////////////////////////////////////
	// Populate the Product struct from the hash
	//////////////////////////////////////////
	product, err := populateProductFromHash(values)
	if err != nil {
		return Product{}, err
	}

	//////////////////////////////////////////
	// Get the product images
	//////////////////////////////////////////
	imageValues, _ := redis.StringMap(redisConn.Do("HGETALL", getProductImagesKeyName(id)))
	product.Images = getProductImagesFromHash(imageValues)

	return product, nil
}
func productExists(id int, redisConn redis.Conn) bool {
	exists, _ := redis.Bool(redisConn.Do("EXISTS", getProductNameById(id)))
	return exists
}

//////////////////////
// IMAGE MODEL
//////////////////////
type Image struct {
	Id        int    `json:"id"`
	Value     []byte `json:"-"`
	ProductId int    `json:"product_id,omitempty"`
	Url       string `json:"url"`
}

func (image *Image) delete(redisConn redis.Conn) {
	_, _ = redisConn.Do("HDEL", config.KeyImages, image.Id)
	_, _ = redisConn.Do("HDEL", getProductImagesKeyName(image.ProductId), image.Id)
	_, _ = redisConn.Do("DEL", getImageNameById(image.Id))
}

func getImageDataById(id int, redisConn redis.Conn) ([]byte, error) {
	return redis.Bytes(redisConn.Do("GET", getImageNameById(id)))
}

func saveImage(productId int, data []byte, redisConn redis.Conn) (Image, error){
	// Get new image id from counter
	imageId := getNextImageId(redisConn)

	// Create image key name
	keyName := getImageNameById(imageId)

	_, err := redisConn.Do("SET", keyName, data)
	if err != nil {
		return Image{}, err
	}

	// Save image to "all images" hash
	_, err = redisConn.Do("HSET", config.KeyImages, imageId, productId)
	if err != nil {
		return Image{}, err
	}

	// Set up the image struct
	image := Image{
		Id:        imageId,
		ProductId: productId,
		Url:       getImageUrlById(imageId),
	}

	// Add image to product's images hash
	productImagesKeyName := getProductImagesKeyName(productId)
	_, err = redisConn.Do("HSET", productImagesKeyName, image.Id, image.Url)
	if err != nil {
		return image, err
	}
	image.Url = config.BaseUri + image.Url

	return image, nil
}

//////////////////////
// CATEGORY MODEL
//////////////////////
type Category struct {
	Id   int    `redis:"id" json:"id"`
	Name string `redis:"name" json:"name"`
}

type PaginatedCollection struct {
	Data           []Product `json:"data"`
	CurrentPage    int       `json:"current_page"`
	ResultsPerPage int       `json:"per_page"`
}
