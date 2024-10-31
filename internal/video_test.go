package internal

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPreviewVideo(t *testing.T) {
	data, err := VideoPreview("../test/sample.mp4", 0)
	assert.NoError(t, err)
	assert.EqualValues(t, 107974, len(data))
	hash, err := SHA256Hash(bytes.NewReader(data))
	assert.NoError(t, err)
	assert.Equal(t, "b29ec13ece50f8343604a26f83216153f3ea58f038d00e1fae9e4461c6c36313", hash)
}
