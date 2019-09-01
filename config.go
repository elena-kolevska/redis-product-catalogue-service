package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	WebServerPort int    `json:"web_server_port"`
	BaseUri       string `json:"base_uri"` // without a trailing slash

	RedisEndpoint string `json:"redis_endpoint"`
	RedisPassword string `json:"redis_password"`

	KeyCategories         string `json:"key_categories"`
	KeyProductCounter     string `json:"key_product_counter"`
	KeyImageCounter       string `json:"key_image_counter"`
	KeyImage              string `json:"key_image"`
	KeyImages             string `json:"key_images"`
	KeyProduct            string `json:"key_product"`
	KeyProductImages      string `json:"key_product_images"`
	KeyAllProducts        string `json:"key_all_products"`
	KeyProductsInCategory string `json:"key_products_in_category"`

	ResultsPerPage int `json:"results_per_page"`
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
		BaseUri:       "http://localhost:8080",

		RedisEndpoint: "localhost:6379",
		RedisPassword: "",

		KeyCategories:         "categories",
		KeyProductCounter:     "product_counter",
		KeyImageCounter:       "image_counter",
		KeyProduct:            "product:%v",
		KeyImage:              "image:%v",
		KeyImages:             "images",
		KeyProductImages:      "product:%v:images",
		KeyAllProducts:        "products",
		KeyProductsInCategory: "products:cat:%v",
		ResultsPerPage:        20,
	}
}
