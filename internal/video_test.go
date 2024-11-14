package internal

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

	preview, err := VideoPreview("../test/sample.mp4", VideoParams{timeout: time.Minute})
	require.NoError(t, err)
	assert.Len(t, preview.Data, 93774)
	assert.EqualValues(t, time.Second*5, preview.Duration)
	assert.Equal(t, "1280x720", preview.Resolution)
	hash := sha256.Sum256(preview.Data)
	assert.Equal(t, "913e1f20eb400f3a13aa043005204ef53e0883c122086b96d94a2b6279ec008e", hex.EncodeToString(hash[:]))
}

func TestPreviewVideoFLV(t *testing.T) {
	t.Parallel()

	preview, err := VideoPreview("../test/video/sample_960x400_ocean_with_audio.flv", VideoParams{timeout: time.Minute})
	require.NoError(t, err)
	assert.Len(t, preview.Data, 11590)
	hash := sha256.Sum256(preview.Data)
	assert.Equal(t, "b7583f7f39807c1ef4636423281f38e6a67d979e3bf2aa0a1a53fb35470c31d1", hex.EncodeToString(hash[:]))
}
