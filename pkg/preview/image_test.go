package preview

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPreviewImage(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	preview, err := imagePreview("../../test/image.jpg")
	require.NoError(t, err)
	assert.Len(preview.Data(), 8620)
	assert.Equal("1527x898", preview.Resolution())
	hash := sha256.Sum256(preview.Data())
	assert.Equal("64de9c944a91c93e750d097577c8fc5992100a7bb186d376534e78705aefbbbd", hex.EncodeToString(hash[:]))
}
