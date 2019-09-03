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


func TestImage_setUrl(t *testing.T) {
	image := Image{
		Id: 78,
	}
	image.setUrl()

	assert.Equal(t, config.BaseUri + fmt.Sprintf("/images/%v", 78), image.Url)
}

func TestImage_delete(t *testing.T) {
	// Too trivial. Should probably remove it
	image := Image{
		Id: 1,
		ProductId: 2,
	}
	conn := redigomock.NewConn()

	cmd1 := conn.Command("MULTI")
	cmd3 := conn.Command("SREM", fmt.Sprintf(config.KeyProductImages, "2"), 1)
	cmd2 := conn.Command("HDEL", config.KeyImages, image.Id)
	cmd4 := conn.Command("DEL", getImageNameById(image.Id))
	cmd5 := conn.Command("EXEC")

	err := image.delete(conn)
	if err != nil {
		t.Error(err)
	}

	if conn.Stats(cmd1) + conn.Stats(cmd2) + conn.Stats(cmd3) + conn.Stats(cmd4) + conn.Stats(cmd5) != 5 {
		t.Error("Some keys weren't deleted properly")
	}
}

func TestSaveNewImage(t *testing.T) {
	imageData := make([]byte, 1*1)

	imageId := 1
	productId := 2

	conn := redigomock.NewConn()
	_ = conn.Command("INCR", config.KeyImageCounter).Expect(int64(imageId))

	_ = conn.Command("SET", getImageNameById(imageId), imageData).Expect("OK")
	_ = conn.Command("HSET", config.KeyImages, imageId, productId).Expect("OK")
	_ = conn.Command("SADD", fmt.Sprintf(config.KeyProductImages, "2"), imageId,).Expect("OK")


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

func TestGetImageNameById(t *testing.T){
	assert.Equal(t, fmt.Sprintf(config.KeyImage, 78), getImageNameById(78))
}
