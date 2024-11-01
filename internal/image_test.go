package internal

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPreviewImage(t *testing.T) {
	previewBytes, err := ImagePreview("../test/image.jpg")
	assert.NoError(t, err)
	assert.EqualValues(t, 18452, len(previewBytes))
	r := bytes.NewReader(previewBytes)
	hash, err := SHA256Hash(r)
	assert.NoError(t, err)
	assert.Equal(t, "029f21ed0e085973dc41f290f5d361185f226c924a7e30e2d2d8c8acac6ade5a", hash)
}
