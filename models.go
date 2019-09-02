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
	Id               int      `redis:"id" json:"id"`
	Name             string   `redis:"name" json:"name"`
	Description      string   `redis:"description" json:"description"`
	Vendor           string   `redis:"vendor" json:"vendor"`
	Price            float32  `redis:"price" json:"price"`
	Currency         string   `redis:"currency" json:"currency"`
	MainCategoryId   int      `redis:"main_category_id" json:"main_category_id,omitempty"`
	MainCategoryName string   `redis:"-" json:"-"`
	MainCategory     Category `redis:"-" json:"main_category"`
	Images           []Image  `redis:"-" json:"images" `
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
func (product *Product) setCategory(redisConn redis.Conn) {
	if product.MainCategoryName == "" {
		categoryName, _ := redis.String(redisConn.Do("HGET", config.KeyCategories, product.MainCategoryId))
		product.MainCategoryName = categoryName
	}
	product.MainCategory = Category{
		Id:   product.MainCategoryId,
		Name: product.MainCategoryName,
	}
	product.MainCategoryId = 0 //We don't want to show this field directly on the product object, but as a part of its category
}
func (product *Product) setImages(redisConn redis.Conn) {
	imageValues, _ := getHashAsStringMap(getProductImagesKeyName(product.Id), redisConn)
	product.Images = getProductImagesFromHash(imageValues)

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
	imageValues, _ := getHashAsStringMap(productImagesKeyName, redisConn)

	// Start a transaction and send all commands in a pipeline
	_, _ = redisConn.Do("MULTI")

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
	product := Product{
		Id: id,
	}

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
	err = redis.ScanStruct(values, &product)
	if err != nil {
		return Product{}, err
	}

	return product, nil
}
func productExists(id int, redisConn redis.Conn) bool {
	exists, _ := redis.Bool(redisConn.Do("EXISTS", getProductNameById(id)))
	return exists
}
func saveNewProduct(product *Product, redisConn redis.Conn) error {
	//////////////////////////////////////////
	// Get a product id from the id counter
	// and assign it to the product struct
	//////////////////////////////////////////
	product.setId()

	/////////////////////
	// Save hash to Redis
	/////////////////////
	_, err := redisConn.Do("HSET", redis.Args{product.getKeyName()}.AddFlat(product)...)
	if err != nil {
		return err
	}

	// Add product to sorted set of all products
	_, err = redisConn.Do("ZADD", config.KeyAllProducts, 0, product.getLexName())
	if err != nil {
		return err
	}

	// Add product to sorted set of products in category
	_, err = redisConn.Do("ZADD", getProductsInCategoryKeyName(product.MainCategoryId), 0, product.getLexName())
	if err != nil {
		return err
	}

	return nil
}

func updateProduct(product *Product, oldProduct *Product, redisConn redis.Conn) error {
	/////////////////////
	// Save hash to Redis
	/////////////////////
	_, err := redisConn.Do("HSET", redis.Args{product.getKeyName()}.AddFlat(product)...)
	if err != nil {
		return err
	}

	//////////////////////////////////////////
	// If the category has been updated remove the product from
	// the old categorised product list and add it to the new one
	//////////////////////////////////////////
	if oldProduct.MainCategoryId != product.MainCategoryId {
		_, err = redisConn.Do("ZREM", getProductsInCategoryKeyName(oldProduct.MainCategoryId), oldProduct.getLexName())
		if err != nil {
			return err
		}

		_, err = redisConn.Do("ZADD", getProductsInCategoryKeyName(product.MainCategoryId), 0, product.getLexName())
		if err != nil {
			return err
		}
	}

	return nil
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

func saveImage(productId int, data []byte, redisConn redis.Conn) (Image, error) {
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
