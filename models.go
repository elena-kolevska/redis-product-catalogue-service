package main

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
)

//////////////////////
// PRODUCT MODEL
//////////////////////
type Product struct {
	Id             int      `redis:"id" json:"id"`
	Name           string   `redis:"name" json:"name"`
	Description    string   `redis:"description" json:"description"`
	Vendor         string   `redis:"vendor" json:"vendor"`
	Price          float32  `redis:"price" json:"price"`
	Currency       string   `redis:"currency" json:"currency"`
	MainCategoryId int      `redis:"main_category_id" json:"main_category_id,omitempty"`
	MainCategory   Category `redis:"-" json:"main_category"`
	Images         []Image  `redis:"-" json:"images" `
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

func (product *Product) setCategory() {
	categoryName, _ := redis.String(redisConn.Do("HGET", config.KeyCategories, product.MainCategoryId))
	category := Category{
		Id:   product.MainCategoryId,
		Name: categoryName,
	}
	product.setCategoryFromStruct(category)
}

func (product *Product) setCategoryFromStruct(category Category) {
	product.MainCategory = category
	product.MainCategoryId = 0 //We don't want to show this field directly on the product object, but as a part of its category
}

//////////////////////
// IMAGE MODEL
//////////////////////
type Image struct {
	Id        int    `json:"id"`
	Value     byte   `json:"-"`
	ProductId int    `json:"product_id"`
	Url       string `json:"url"`
}

//////////////////////
// CATEGORY MODEL
//////////////////////
type Category struct {
	Id   int    `redis:"id" json:"id"`
	Name string `redis:"name" json:"name"`
}
