package internal

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSHA256Hash(t *testing.T) {
	// img
	f, err := os.Open("../test/image.jpg")
	assert.NoError(t, err)
	hash, err := SHA256Hash(f)
	assert.NoError(t, err)
	assert.Equal(t, "9aaf695847bad105d79aaccd448562fb4cb8a2e64798e9611a012c7c946c4f57", hash)
	// video
	f, err = os.Open("../test/sample.mp4")
	assert.NoError(t, err)
	hash, err = SHA256Hash(f)
	assert.NoError(t, err)
	assert.Equal(t, "f25b31f155970c46300934bda4a76cd2f581acab45c49762832ffdfddbcf9fdd", hash)
}
