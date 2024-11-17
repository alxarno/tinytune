package preview

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPreviewVideoFLV(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	preview, err := videoPreview("../../test/video/sample_960x400_ocean_with_audio.flv", VideoParams{timeout: time.Minute})
	require.NoError(t, err)
	assert.Len(preview.Data(), 11590)
	hash := sha256.Sum256(preview.Data())
	assert.Equal("b7583f7f39807c1ef4636423281f38e6a67d979e3bf2aa0a1a53fb35470c31d1", hex.EncodeToString(hash[:]))
}
