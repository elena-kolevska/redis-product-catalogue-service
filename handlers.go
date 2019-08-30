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
	categoryName, err := redis.String(redisConn.Do("HGET", "categories", product.MainCategoryId))
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
	_, _ = redisConn.Do("ZADD", "products", product.Price, product.Id)

	// Add product to sorted set of products in category
	productsInCategoryKeyName := fmt.Sprintf("products:cat:%v", product.MainCategoryId)
	_, _ = redisConn.Do("ZADD", productsInCategoryKeyName, product.Price, product.Id)

	// Add product to sorted set for lexicographical sorting (prefix searching)
	_, _ = redisConn.Do("ZADD", "products:lex", 0, product.getLexName())

	// Add product to sorted set of products in category for lexicographical sorting (prefix searching)
	productsInCategoryLexKeyName := fmt.Sprintf("products:lex:cat:%v", product.MainCategoryId)
	_, _ = redisConn.Do("ZADD", productsInCategoryLexKeyName, 0, product.getLexName())

	category := Category{
		Id:   product.MainCategoryId,
		Name: categoryName,
	}
	// Format for presentation
	product.setCategory(category)

	return c.JSON(http.StatusCreated, product)
}

func productsIndex(c echo.Context) error {
	return c.String(http.StatusOK, "Products Index")
}

func productsShow(c echo.Context) error {
	return c.String(http.StatusOK, "Products Show")
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
