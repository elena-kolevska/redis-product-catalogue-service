package main

import (
	"fmt"
	"github.com/rafaeljusto/redigomock"
	"gotest.tools/assert"
	"testing"
)



func TestImage_setId(t *testing.T) {
	image := Image{}
	conn := redigomock.NewConn()
	cmd := conn.Command("INCR", config.KeyImageCounter).Expect(int64(1))
	image.setId(conn)

	if conn.Stats(cmd) != 1 {
		t.Error("INCR Call to Redis was never made")
	}
	assert.Equal(t, 1, image.Id)
}

func TestImage_delete(t *testing.T) {
	// Too trivial. Should probably remove it

	image := Image{
		Id: 1,
	}
	conn := redigomock.NewConn()

	cmd1 := conn.Command("HDEL", config.KeyImages, image.Id)
	cmd2 := conn.Command("HDEL", getProductImagesKeyName(image.ProductId), image.Id)
	cmd3 := conn.Command("DEL", getImageNameById(image.Id))

	image.delete(conn)

	if conn.Stats(cmd1) + conn.Stats(cmd2) + conn.Stats(cmd3) != 3 {
		t.Error("Keys weren't deleted properly")
	}
}

func TestSaveNewImage(t *testing.T) {
	imageData := make([]byte, 1*1)

	imageId := 1
	productId := 2

	conn := redigomock.NewConn()
	_ = conn.Command("INCR", config.KeyImageCounter).Expect(int64(imageId))
	keyName := getImageNameById(imageId)

	_ = conn.Command("SET", keyName, imageData).Expect("OK")
	_ = conn.Command("HSET", config.KeyImages, imageId, productId).Expect("OK")
	_ = conn.Command("HSET", fmt.Sprintf(config.KeyProductImages, "2"), imageId, config.BaseUri + "/images/1",).Expect("OK")


	image, err := saveNewImage(productId, imageData, conn)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, image, Image{
		Id:        imageId,
		ProductId: productId,
		Url:       config.BaseUri + "/images/1",
	})
}

