package main

import (
	"fmt"
	"github.com/bugsnag/bugsnag-go"
	"github.com/gomodule/redigo/redis"
	"github.com/labstack/echo"
	"github.com/labstack/gommon/log"
	"io/ioutil"
	"net/http"
	"strconv"
)

func productsCreate(c echo.Context) error {
	product := Product{}

	//////////////////////////////////////////
	// Read the json data into `product` and implicitly check if all data types are correct
	// (ex. can't send a string as price or category id)
	//////////////////////////////////////////
	if err := c.Bind(&product); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, validationError)
	}

	//////////////////////////////////////////
	// Check presence of required fields
	// TODO Confirm this is the only required field
	//////////////////////////////////////////
	if product.Name == "" {
		return c.JSON(http.StatusUnprocessableEntity, ApiError{Title: "The name field is required", Description: "Please provide a product name"})
	}

	//////////////////////////////////////////
	// Check category id exists
	//////////////////////////////////////////
	categoryName, err := getCategoryNameById(product.MainCategoryId, redisConn)
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, ApiError{Title: "Category doesn't exist", Description: "That category id doesn't exist in our system"})
	}
	product.MainCategoryName = categoryName

	err = saveNewProduct(&product, redisConn)
	if err != nil {
		return serverErrorResponse(c, err)
	}

	product.setCategory(redisConn)
	return c.JSON(http.StatusCreated, product)
}

func productsIndex(c echo.Context) error {

	var command string
	args:= redis.Args{}
	keyName := config.KeyAllProducts
	categories := getCategoriesMap(redisConn)

	////////////////////////////////////////////////////
	// Check if we need to show all products or only products in a certain category
	////////////////////////////////////////////////////
	mainCategoryIdParam := c.QueryParam("main_category_id")
	if len(mainCategoryIdParam) > 0 {
		mainCategoryId, _ := strconv.Atoi(mainCategoryIdParam)

		// Check if category id exists and if it does, look into a different key (products by category)
		_, ok := categories[mainCategoryId]
		if ok {
			keyName = fmt.Sprintf(config.KeyProductsInCategory, mainCategoryId)
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
		command = "ZRANGEBYLEX"
		args = args.Add(keyName).Add(fromArg).Add(toArg).Add("LIMIT").Add(fromPosition).Add(config.ResultsPerPage)
	} else {
		command = "ZRANGE"
		args = args.Add(keyName).Add(fromPosition).Add(toPosition)
	}

	products, err := getProducts(command, args, categories, redisConn)
	if err != nil {
		return serverErrorResponse(c, err)
	}

	response := PaginatedProductCollection{
		CurrentPage:    pageNumber,
		ResultsPerPage: config.ResultsPerPage,
		Data:           products,
	}
	return c.JSON(http.StatusOK, response)
}

func productsShow(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(urlParamError.HttpStatus, urlParamError)
	}

	product, err := getProductById(id, redisConn)
	if err != nil {
		switch e := err.(type) {
		case *ApiError:
			return c.JSON(e.HttpStatus, e)
		default:
			return serverErrorResponse(c, err)
		}
	}
	product.setCategory(redisConn)
	product.setImages(redisConn)

	return c.JSON(http.StatusOK, product)
}

func productsUpdate(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(urlParamError.HttpStatus, urlParamError)
	}

	oldProduct, err := getProductById(id, redisConn)
	if err != nil {
		switch e := err.(type) {
		case *ApiError:
			return c.JSON(e.HttpStatus, e)
		default:
			return serverErrorResponse(c, err)
		}
	}

	product := Product{
		Id: id,
	}
	//////////////////////////////////////////
	// Read the json data into `product` and implicitly check if all data types are correct
	// (ex. can't send a string as price or category id)
	//////////////////////////////////////////
	if err := c.Bind(&product); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, validationError)
	}

	//////////////////////////////////////////
	// Check presence of required fields
	// TODO Confirm this is the only required field
	//////////////////////////////////////////
	if product.Name == "" {
		return c.JSON(http.StatusUnprocessableEntity, ApiError{Title: "The name field is required", Description: "Please provide a product name"})
	}

	//////////////////////////////////////////
	// Get category name
	//////////////////////////////////////////
	categoryName, err := getCategoryNameById(product.MainCategoryId, redisConn)
	product.MainCategoryName = categoryName
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, ApiError{Title: "Category doesn't exist", Description: "That category id doesn't exist in our system"})
	}


	err = updateProduct(&product, &oldProduct, redisConn)
	if err != nil {
		return serverErrorResponse(c, err)
	}

	product.setCategory(redisConn)
	product.setImages(redisConn)

	return c.JSON(http.StatusOK, product)
}

func productsDelete(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(urlParamError.HttpStatus, urlParamError)
	}

	product := Product{
		Id: id,
	}

	err = product.delete(redisConn)
	if err != nil {
		switch e := err.(type) {
		case *ApiError:
			return c.JSON(e.HttpStatus, e)
		default:
			return serverErrorResponse(c, err)
		}
	}

	return c.NoContent(http.StatusNoContent)
}

func imagesShow(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, urlParamError)
	}

	data, err := getImageDataById(id, redisConn)
	if err != nil {
		return c.JSON(notFoundError.HttpStatus, notFoundError)
	}

	// We can use a jpg subtype here, even if the image is of a different type. The browser still renders it properly.
	return c.Blob(http.StatusCreated, "image/jpg", data)
}

func imagesCreate(c echo.Context) error {
	productId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, urlParamError)
	}

	// Check if product exists
	if productExists(productId, redisConn) == false {
		return c.JSON(notFoundError.HttpStatus, notFoundError)
	}

	body, _ := ioutil.ReadAll(c.Request().Body)

	image, err := saveNewImage(productId, body, redisConn)
	if err != nil {
		return serverErrorResponse(c, err)
	}

	return c.JSON(http.StatusCreated, image)
}

func imagesDelete(c echo.Context) error {
	imageId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, urlParamError)
	}

	productId, err := redis.Int(redisConn.Do("HGET", config.KeyImages, imageId))
	if err == redis.ErrNil {
		return c.JSON(http.StatusNotFound, notFoundError)
	}

	image := Image{
		Id:        imageId,
		ProductId: productId,
	}
	err = image.delete(redisConn)
	if err != nil {
		return serverErrorResponse(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}

func serverErrorResponse(c echo.Context, err error) error {
	log.Error(err)
	_ = bugsnag.Notify(err)
	return c.JSON(serverError.HttpStatus, serverError)
}
