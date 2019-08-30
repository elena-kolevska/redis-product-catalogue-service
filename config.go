package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	WebServerPort            int    `json:"web_server_port"`

	RedisEndpoint            string `json:"redis_endpoint"`
	RedisPassword            string `json:"redis_password"`

	KeyCategories            string `json:"key_categories"`
	KeyProductCounter        string `json:"key_product_counter"`
	KeyProduct               string `json:"key_product"`
	KeyAllProducts           string `json:"key_all_products"`
	KeyAllProductsLex        string `json:"key_all_products_lex"`
	KeyProductsInCategory    string `json:"key_products_in_category"`
	KeyProductsInCategoryLex string `json:"key_products_in_category_lex"`
}

func getConfiguration() Config {
	// Get configuration values
	configFile, _ := os.Open("conf.json")
	defer configFile.Close()

	decoder := json.NewDecoder(configFile)
	config := Config{}
	err := decoder.Decode(&config)

	// If there's an error in the json config file we resort to default values
	if err != nil {
		fmt.Println("❌ Config file error: ", err)
		fmt.Println("Reading default configuration.")
		config = getDefaultConfiguration()
	}

	return config
}

func getDefaultConfiguration() Config {
	return Config{
		WebServerPort: 8080,

		RedisEndpoint: "localhost:6379",
		RedisPassword: "",

		KeyCategories:            "categories",
		KeyProductCounter:        "product_counter",
		KeyProduct:               "product:%v",
		KeyAllProducts:           "products",
		KeyAllProductsLex:        "products:lex",
		KeyProductsInCategory:    "products:cat:%v",
		KeyProductsInCategoryLex: "products:lex:cat:%v",
	}
}
