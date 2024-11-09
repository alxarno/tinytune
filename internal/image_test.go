package internal

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPreviewImage(t *testing.T) {
	preview, err := ImagePreview("../test/image.jpg")
	assert.NoError(t, err)
	assert.EqualValues(t, 8620, len(preview.Data))
	assert.Equal(t, "1527x898", preview.Resolution)
	r := bytes.NewReader(preview.Data)
	hash, err := SHA256Hash(r)
	assert.NoError(t, err)
	assert.Equal(t, "64de9c944a91c93e750d097577c8fc5992100a7bb186d376534e78705aefbbbd", hash)
}
