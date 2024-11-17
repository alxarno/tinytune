package preview

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPreviewVideo(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	preview, err := videoPreview("../../test/sample.mp4", VideoParams{timeout: time.Minute})
	require.NoError(t, err)
	assert.Len(preview.Data(), 107832)
	assert.EqualValues(time.Second*5, preview.Duration())
	width, height := preview.Resolution()
	assert.Equal(1280, width)
	assert.Equal(720, height)

	hash := sha256.Sum256(preview.Data())
	assert.Equal("7f2d1e244e636db9178a8cb49d375ce3268048fddd9efa1d7943e03003a2b708", hex.EncodeToString(hash[:]))
}

func TestPreviewVideoFLV(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	preview, err := videoPreview("../../test/video/sample_960x400_ocean_with_audio.flv", VideoParams{timeout: time.Minute})
	require.NoError(t, err)
	assert.Len(preview.Data(), 37038)

	hash := sha256.Sum256(preview.Data())
	assert.Equal("89f3a6b1f30ef25f1e4b1fc4cdfe6af7c7d9632d58c2ef15bdb88d7302ced91f", hex.EncodeToString(hash[:]))
}

func TestPreviewShortVideo(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	preview, err := videoPreview("../../test/short.mp4", VideoParams{timeout: time.Minute})
	require.NoError(t, err)
	assert.Len(preview.Data(), 107340)

	hash := sha256.Sum256(preview.Data())
	assert.Equal("43d96ebec876b891810f55bb8e99529ea52ebc325d4ed1d0113ab85afc115a95", hex.EncodeToString(hash[:]))
}
