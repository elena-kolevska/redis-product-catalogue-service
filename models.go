package main

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"strings"
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
	id, _ := redis.Int(redisConn.Do("INCR", "product_counter"))
	product.Id = id
}
func (product *Product) getKeyName() string {
	return fmt.Sprintf("product:%v", product.Id)
}
func (product *Product) getLexName() string {
	return fmt.Sprintf("%s::%v", product.getNormalisedName(), product.Id)
}
func (product *Product) getNormalisedName() string {
	return strings.Replace(strings.ToLower(product.Name), "::", " ", -1)
}

func (product *Product) setCategory(category Category) {
	product.MainCategory = category
	product.MainCategoryId = 0 //We don't want to show this field directly on the product object, but as a part of its category
}

//////////////////////
// IMAGE MODEL
//////////////////////
type Image struct {
	Id        int  `redis:"id" json:"id"`
	Value     byte `redis:"value" json:"value"`
	ProductId int  `redis:"value" json:"value"`
}

//////////////////////
// CATEGORY MODEL
//////////////////////
type Category struct {
	Id   int    `redis:"id" json:"id"`
	Name string `redis:"name" json:"name"`
}
