package internal

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPreviewVideo(t *testing.T) {
	data, err := VideoPreview("../test/sample.mp4", videoParams{timeout: time.Minute})
	assert.NoError(t, err)
	assert.EqualValues(t, 93774, len(data))
	hash, err := SHA256Hash(bytes.NewReader(data))
	assert.NoError(t, err)
	assert.Equal(t, "913e1f20eb400f3a13aa043005204ef53e0883c122086b96d94a2b6279ec008e", hash)
}
