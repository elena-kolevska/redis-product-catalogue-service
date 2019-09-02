package main

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"github.com/labstack/gommon/log"
	"strconv"
	"strings"
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
	values, _ := getHashAsStringMap(getProductImagesKeyName(product.Id), redisConn)
	product.Images = getProductImagesFromHash(values)
}

// Similar behavior as `setImages` but uses the string map received from a pipeline
func (product *Product) setImagesFromStringMap(values map[string]string) {
	product.Images = getProductImagesFromHash(values)
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

func getProductImagesFromHash(values map[string]string) []Image {
	images := make([]Image, 0)

	for imageId, imageUrl := range values {
		imageId, _ := strconv.Atoi(imageId)
		image := Image{
			Id:  imageId,
			Url: config.BaseUri + imageUrl,
		}
		images = append(images, image)
	}

	return images
}

// Helper functions
func getProductNameById(id int) string {
	return fmt.Sprintf(config.KeyProduct, id)
}
func getProductImagesKeyName(id int) string {
	return fmt.Sprintf(config.KeyProductImages, strconv.Itoa(id))
}
func getProductsInCategoryKeyName(categoryId int) string {
	return fmt.Sprintf(config.KeyProductsInCategory, categoryId)
}

type PaginatedProductCollection struct {
	Data           []Product `json:"data"`
	CurrentPage    int       `json:"current_page"`
	ResultsPerPage int       `json:"per_page"`
}

func getProducts(command string, args redis.Args, categories map[int]Category, redisConn redis.Conn) ([]Product,error) {

	products := make([]Product, 0)

	results, err := redis.Strings(redisConn.Do(command, args...))
	if err != nil {
		return products, err
	}
	////////////////////////////////////////////////////
	// If no results - respond with an empty json array
	////////////////////////////////////////////////////
	if len(results) == 0 {
		return products, nil
	}

	////////////////////////////////////////////////////
	// Send all the HGETALL commands in a pipeline, so we don't need to make too many requests to the database
	////////////////////////////////////////////////////
	for _, product := range results {
		temp := strings.Split(product, "::")
		productId, _ := strconv.Atoi(temp[1])

		// Get the product data
		err := redisConn.Send("HGETALL", getProductNameById(productId))
		if err != nil {
			return products, nil
		}
		// Get the product images
		err = redisConn.Send("HGETALL", getProductImagesKeyName(productId))
		if err != nil {
			return products, nil
		}
	}

	_ = redisConn.Flush()

	////////////////////////////////////////////////////
	// Call "Receive" on the client for every hash in the collection,
	// scan it into a struct and append it into the resulting collection
	////////////////////////////////////////////////////
	for _, _ = range results {
		values, _ := redis.Values(redisConn.Receive())

		var product Product
		_ = redis.ScanStruct(values, &product)
		product.MainCategory = categories[product.MainCategoryId]
		product.MainCategoryId = 0

		// Now grab the image data
		imageValues, _ := redis.StringMap(redisConn.Receive())
		product.setImagesFromStringMap(imageValues)

		products = append(products, product)
	}

	return products, nil
}