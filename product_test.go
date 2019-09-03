package main

import (
	"fmt"
	"github.com/rafaeljusto/redigomock"
	"gotest.tools/assert"
	"testing"
)

func TestProduct_setId(t *testing.T) {
	product := Product{}
	conn := redigomock.NewConn()
	cmd := conn.Command("INCR", config.KeyProductCounter).Expect(int64(78))
	product.setId(conn)

	if conn.Stats(cmd) != 1 {
		t.Error("INCR Call to Redis was never made")
	}
	assert.Equal(t, 78, product.Id)
}

func TestProduct_getKeyName(t *testing.T) {
	product := Product{Id: 78}

	assert.Equal(t, fmt.Sprintf(config.KeyProduct, 78), product.getKeyName())
}
func TestProduct_getLexName(t *testing.T) {
	product := Product{Id: 78, Name: "Product Name"}

	assert.Equal(t, "product name::78", product.getLexName())
}

func TestNormaliseSearchString(t *testing.T) {
	testCases := [] struct {
		arg string
		want string
	}{
		{"product name", "product name"},
		{"product :: Name", "product   name"},
		{"Product :: Name", "product   name"},
		{"Product Name", "product name"},
		{"PRODUCT NAME", "product name"},
	}

	for _,tc := range testCases {
		got := normaliseSearchString(tc.arg)
		if got != tc.want {
			t.Errorf("Normalised version of '%s' should be '%s' (returned '%s')", tc.arg, tc.want, got)
		}
	}

}

func TestGetProductById(t *testing.T) {

}
