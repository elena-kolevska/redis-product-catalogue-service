package main

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"github.com/labstack/echo"
	"github.com/labstack/gommon/log"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

func productsCreate(c echo.Context) error {
	product := new(Product)

	//////////////////////////////////////////
	// Implicitly check if all data types are correct
	// (ex. can't send a string as price or category id)
	//////////////////////////////////////////
	if err := c.Bind(product); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, validationError)
	}

	//////////////////////////////////////////
	// Check presence of required fields
	// TODO Confirm this is the only required field
	//////////////////////////////////////////
	if product.Name == "" {
		return c.JSON(http.StatusUnprocessableEntity, Error{Title: "The name field is required", Description: "Please provide a product name"})
	}

	//////////////////////////////////////////
	// Check if category id exists
	//////////////////////////////////////////
	categoryName, err := getCategoryNameById(product.MainCategoryId)
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, Error{Title: "Category doesn't exist", Description: "That category id doesn't exist in our system"})
	}

	//////////////////////////////////////////
	// Get a product id from the id counter
	// and assign it to the product struct
	//////////////////////////////////////////
	product.setId()

	/////////////////////
	// Save hash to Redis
	/////////////////////
	_, err = redisConn.Do("HSET", redis.Args{product.getKeyName()}.AddFlat(product)...)
	if err != nil {
		return serverErrorResponse(c, err)
	}

	// Add product to sorted set of all products
	_, _ = redisConn.Do("ZADD", config.KeyAllProducts, product.Price, product.Id)

	// Add product to sorted set of products in category
	productsInCategoryKeyName := fmt.Sprintf(config.KeyProductsInCategory, product.MainCategoryId)
	_, _ = redisConn.Do("ZADD", productsInCategoryKeyName, product.Price, product.Id)

	// Add product to sorted set for lexicographical sorting (prefix searching)
	_, _ = redisConn.Do("ZADD", config.KeyAllProductsLex, 0, product.getLexName())

	// Add product to sorted set of products in category for lexicographical sorting (prefix searching)
	productsInCategoryLexKeyName := fmt.Sprintf(config.KeyProductsInCategoryLex, product.MainCategoryId)
	_, _ = redisConn.Do("ZADD", productsInCategoryLexKeyName, 0, product.getLexName())

	category := Category{
		Id:   product.MainCategoryId,
		Name: categoryName,
	}
	// Format for presentation
	product.setCategoryFromStruct(category)

	return c.JSON(http.StatusCreated, product)
}

func productsIndex(c echo.Context) error {

	var results []string
	keyName := config.KeyAllProductsLex
	products := make([]Product, 0)
	categories := getCategoriesMap()

	////////////////////////////////////////////////////
	// Check if we need to show all products or only products in a certain category
	////////////////////////////////////////////////////
	mainCategoryIdParam := c.QueryParam("main_category_id")
	if len(mainCategoryIdParam) > 0 {
		mainCategoryId, _ := strconv.Atoi(mainCategoryIdParam)

		// Check if category id exists and if it does, look into a different key (products by category)
		_, ok := categories[mainCategoryId]
		if ok {
			keyName = fmt.Sprintf(config.KeyProductsInCategoryLex, mainCategoryId)
		}
	}

	////////////////////////////////////////////////////
	// Get pagination positions
	////////////////////////////////////////////////////
	pageNumber, _ := strconv.Atoi(c.QueryParam("page"))
	if pageNumber < 1 {
		pageNumber = 1
	}
	fromPosition := (pageNumber - 1) * config.ResultsPerPage
	toPosition := fromPosition + config.ResultsPerPage - 1

	////////////////////////////////////////////////////
	// Check if we need to search by name (prefix)
	////////////////////////////////////////////////////
	if c.QueryParam("search") != "" {
		searchString := normaliseSearchString(c.QueryParam("search"))
		fromArg := "[" + searchString
		toArg := "[" + searchString + "\xff"
		results, _ = redis.Strings(redisConn.Do("ZRANGEBYLEX", keyName, fromArg, toArg, "LIMIT", fromPosition, config.ResultsPerPage))
	} else {
		results, _ = redis.Strings(redisConn.Do("ZRANGE", keyName, fromPosition, toPosition))
	}

	////////////////////////////////////////////////////
	// If no results - respond with an empty json array
	////////////////////////////////////////////////////
	if len(results) == 0 {
		return c.JSON(http.StatusOK, products)
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
			return serverErrorResponse(c, err)
		}
		// Get the product images
		err = redisConn.Send("HGETALL", getProductImagesKeyName(productId))
		if err != nil {
			return serverErrorResponse(c, err)
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

		// Now grab the image data
		imageValues, _ := redis.StringMap(redisConn.Receive())
		product.Images = getProductImagesFromHash(imageValues)

		products = append(products, product)
	}

	return c.JSON(http.StatusOK, products)
}

func productsShow(c echo.Context) error {
	id := c.Param("id")
	productKeyName := fmt.Sprintf(config.KeyProduct, id)

	//////////////////////////////////////////
	// Fetch the details of a specific product.
	//////////////////////////////////////////
	values, err := redis.Values(redisConn.Do("HGETALL", productKeyName))
	if err != nil {
		return serverErrorResponse(c, err)
	}
	// If no product is found for the given id, return a 404
	if len(values) == 0 {
		return c.JSON(http.StatusNotFound, notFoundError)
	}

	//////////////////////////////////////////
	// Populate the Product struct from the hash
	//////////////////////////////////////////
	product, err := populateProductFromHash(values)
	if err != nil {
		return serverErrorResponse(c, err)
	}

	return c.JSON(http.StatusOK, product)
}

func productsUpdate(c echo.Context) error {
	return c.String(http.StatusOK, "Products Update")
}

func productsDelete(c echo.Context) error {
	return c.NoContent(http.StatusNoContent)
}

func imagesGet(c echo.Context) error {
	imageId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, Error{
			Title:       "Wrong parameters",
			Description: "The image id in the url needs to be a valid number",
		})
	}

	data,err := redis.Bytes(redisConn.Do("GET", getImageNameById(imageId)))
	if err != nil {
		return c.JSON(http.StatusNotFound, notFoundError)
	}

	return c.Blob(http.StatusCreated, "image/jpg", data)
}
func imagesCreate(c echo.Context) error {

	productId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, Error{
			Title:       "Wrong parameters",
			Description: "The product id in the url needs to be a valid number",
		})
	}
	// TODO Check if product exists

	// Get new image id from counter
	imageId, err := redis.Int(redisConn.Do("INCR", config.KeyProductCounter))
	if err != nil {
		return serverErrorResponse(c, err)
	}

	// Create image key name
	keyName := getImageNameById(imageId)


	// Save image to Redis
	body, _ := ioutil.ReadAll(c.Request().Body)
	_, _ = redisConn.Do("SET", keyName, body)

	// Set up the image struct
	image := Image{
		Id:        imageId,
		ProductId: productId,
		Url:       getImageUrlById(imageId),
	}

	// Add image to product's images hash
	productImagesKeyName := getProductImagesKeyName(productId)
	_ ,_ = redisConn.Do("HSET", productImagesKeyName, image.Id, image.Url)

	return c.JSON(http.StatusCreated, image)
}
func imagesDelete(c echo.Context) error {
	return c.String(http.StatusNoContent, "Images Delete")
}

func serverErrorResponse(c echo.Context, err error) error {
	log.Error(err)
	return c.JSON(http.StatusInternalServerError, serverError)
}
