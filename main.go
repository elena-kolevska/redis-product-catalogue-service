package main

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"github.com/labstack/echo"
	"log"
)

var (
	pool      = newPool()
	redisConn = pool.Get()
)

func main() {
	defer redisConn.Close()

	e := echo.New()
	e.HideBanner = true

	// Register routes
	e.GET("/products", productsIndex)
	e.POST("/products", productsCreate)
	e.GET("/products/:id", productsShow)
	e.PATCH("/products/:id", productsUpdate)
	e.DELETE("/products/:id", productsDelete)

	e.POST("/products/:id/images", imagesCreate)
	e.GET("/images/:id", imagesGet)

	e.File("/documentation", "documentation/index.html")

	// Start the server
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", 8080)))
}

func newPool() *redis.Pool {
	log.Println("Connecting to Redis...")
	return &redis.Pool{
		MaxIdle:   20,
		MaxActive: 1000, // max number of connections
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", "localhost:6379")
			if err != nil {
				panic(err.Error())
			}
			return c, err
		},
	}
}