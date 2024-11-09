package internal

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPreviewVideo(t *testing.T) {
	preview, err := VideoPreview("../test/sample.mp4", videoParams{timeout: time.Minute})
	assert.NoError(t, err)
	assert.EqualValues(t, 93774, len(preview.Data))
	assert.EqualValues(t, time.Second*5, preview.Duration)
	assert.Equal(t, "1280x720", preview.Resolution)
	hash, err := SHA256Hash(bytes.NewReader(preview.Data))
	assert.NoError(t, err)
	assert.Equal(t, "913e1f20eb400f3a13aa043005204ef53e0883c122086b96d94a2b6279ec008e", hash)
}
