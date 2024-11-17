package preview

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
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
	assert.Len(preview.Data(), 105634, fmt.Sprintf("got %d", len(preview.Data())))
	assert.EqualValues(time.Second*5, preview.Duration())
	assert.Equal("1280x720", preview.Resolution())
	hash := sha256.Sum256(preview.Data())
	assert.Equal("ab9e213afa42466148583b1ddab5bc00f8218ecac230cc7f8b6797e9e5179ba2", hex.EncodeToString(hash[:]))
}

func TestPreviewVideoFLV(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	preview, err := videoPreview("../../test/video/sample_960x400_ocean_with_audio.flv", VideoParams{timeout: time.Minute})
	require.NoError(t, err)
	assert.Len(preview.Data(), 37618, fmt.Sprintf("got %d", len(preview.Data())))
	hash := sha256.Sum256(preview.Data())
	assert.Equal("30d4dcd9d15d17939a74bfbc7c4a19126a61369504d308bc11847f1bb45e4627", hex.EncodeToString(hash[:]))
}
