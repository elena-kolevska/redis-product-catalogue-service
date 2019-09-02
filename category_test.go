package main

import (
	"github.com/rafaeljusto/redigomock"
	"gotest.tools/assert"
	"testing"
)

func TestGetCategoriesMap(t *testing.T) {
	conn := redigomock.NewConn()

	cmd := conn.Command("HGETALL", config.KeyCategories).ExpectMap(map[string]string{
		"10": "Cat 1",
		"20": "Cat 2",
	})

	categories := getCategoriesMap(conn)

	if conn.Stats(cmd) != 1 {
		t.Error("HGETALL Call to Redis wasn't made")
	}

	assert.Equal(t, categories[10], Category{
		Id:   10,
		Name: "Cat 1",
	})
	assert.Equal(t, categories[20], Category{
		Id:   20,
		Name: "Cat 2",
	})
}