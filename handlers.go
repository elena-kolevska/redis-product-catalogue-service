package main

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"github.com/labstack/echo"
	"github.com/labstack/gommon/log"
	"net/http"
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
	categoryName, err := redis.String(redisConn.Do("HGET", config.KeyCategories, product.MainCategoryId))
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
	return c.String(http.StatusOK, "Products Index")
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
	var product Product
	err = redis.ScanStruct(values, &product)
	if err != nil {
		return serverErrorResponse(c, err)
	}

	//////////////////////////////////////////
	// Get the product category and attach it to the product struct
	//////////////////////////////////////////
	product.setCategory()

	return c.JSON(http.StatusOK, product)
}

func productsUpdate(c echo.Context) error {
	return c.String(http.StatusOK, "Products Update")
}

func productsDelete(c echo.Context) error {
	return c.NoContent(http.StatusNoContent)
}

func imagesGet(c echo.Context) error {
	return c.String(http.StatusOK, "Images Get")
}
func imagesCreate(c echo.Context) error {
	return c.String(http.StatusCreated, "Images Create")
}
func imagesDelete(c echo.Context) error {
	return c.String(http.StatusNoContent, "Images Delete")
}

func serverErrorResponse(c echo.Context, err error) error {
	log.Error(err)
	return c.JSON(http.StatusInternalServerError, serverError)
}
