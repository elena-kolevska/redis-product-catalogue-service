package main

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	config = getDefaultConfiguration()
	os.Exit(m.Run())
}
