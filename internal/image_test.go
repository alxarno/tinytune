package internal

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPreviewImage(t *testing.T) {
	t.Parallel()

	preview, err := ImagePreview("../test/image.jpg")
	require.NoError(t, err)
	assert.Len(t, preview.Data, 8620)
	assert.Equal(t, "1527x898", preview.Resolution)
	hash := sha256.Sum256(preview.Data)
	assert.Equal(t, "64de9c944a91c93e750d097577c8fc5992100a7bb186d376534e78705aefbbbd", hex.EncodeToString(hash[:]))
}
