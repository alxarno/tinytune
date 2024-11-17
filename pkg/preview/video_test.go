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
	assert.Len(preview.Data(), 93774)
	assert.EqualValues(time.Second*5, preview.Duration())
	assert.Equal("1280x720", preview.Resolution())
	hash := sha256.Sum256(preview.Data())
	assert.Equal("913e1f20eb400f3a13aa043005204ef53e0883c122086b96d94a2b6279ec008e", hex.EncodeToString(hash[:]))
}
