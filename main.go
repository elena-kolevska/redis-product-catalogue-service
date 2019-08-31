package main

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"github.com/labstack/echo"
)

var (
	pool      = newPool()
	redisConn = pool.Get()
	config    = getConfiguration()
)

func main() {
	defer redisConn.Close()

	// Authenticate with Redis if a password was provided in the conf file
	if len(config.RedisPassword) > 0 {
		_, err := redisConn.Do("AUTH", config.RedisPassword)
		if err != nil {
			fmt.Println("❌ Unable to authenticate with the Redis database. Please check your settings in the config.json file")
			panic(err)
		}
		fmt.Println("🔑️ Authenticated with Redis...")
	}
	seedDatabase()


	e := echo.New()
	e.HideBanner = true

	// Register routes
	e.POST("/products", productsCreate)
	e.GET("/products", productsIndex)
	e.GET("/products/:id", productsShow)
	e.PATCH("/products/:id", productsUpdate)
	e.DELETE("/products/:id", productsDelete)

	e.POST("/products/:id/images", imagesCreate)
	e.GET("/images/:id", imagesGet)
	e.DELETE("/images/:id", imagesDelete)

	e.File("/documentation", "documentation/index.html")

	// Start the server
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", config.WebServerPort)))
}

func newPool() *redis.Pool {
	fmt.Println("❤️  Connecting to Redis...")
	return &redis.Pool{
		MaxIdle:   20,
		MaxActive: 1000, // max number of connections
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", config.RedisEndpoint, )
			if err != nil {
				panic(err.Error())
			}
			return c, err
		},
	}
}


func seedDatabase() {
	// In our exercise the API consumer is not able to manage categories
	// so to keep things simple we will use hardcoded category ids
	// instead of counter id generators
	_, _ = redisConn.Do("HSET", "categories", "1", "Science vessels", "2", "Warships", "3", "Freighters", "4", "Colony Ships")
}