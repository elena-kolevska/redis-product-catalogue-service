package main

import (
	"github.com/labstack/echo"
	"net/http"
)

func productsIndex(c echo.Context) error {
	return c.String(http.StatusOK, "Products Index")
}

func productsCreate(c echo.Context) error {
	return c.String(http.StatusCreated, "Products Create")
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
