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

	cases := []struct {
		Name       string
		SourcePath string
		DataLength int
		Duration   time.Duration
		Width      int
		Height     int
		Hash       string
	}{
		{
			Name:       ".mp4",
			SourcePath: "../../test/sample.mp4",
			DataLength: 107832,
			Duration:   time.Second * 5,
			Width:      1280,
			Height:     720,
			Hash:       "7f2d1e244e636db9178a8cb49d375ce3268048fddd9efa1d7943e03003a2b708",
		},
		{
			Name:       ".flv",
			SourcePath: "../../test/video/sample_960x400_ocean_with_audio.flv",
			DataLength: 37038,
			Duration:   time.Second * 46,
			Width:      960,
			Height:     400,
			Hash:       "89f3a6b1f30ef25f1e4b1fc4cdfe6af7c7d9632d58c2ef15bdb88d7302ced91f",
		},
		{
			Name:       "short",
			SourcePath: "../../test/short.mp4",
			DataLength: 107340,
			Duration:   time.Second * 3,
			Width:      1280,
			Height:     720,
			Hash:       "43d96ebec876b891810f55bb8e99529ea52ebc325d4ed1d0113ab85afc115a95",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()

			assert := assert.New(t)
			preview, err := videoPreview(testCase.SourcePath, VideoParams{timeout: time.Minute})
			require.NoError(t, err)
			assert.Len(preview.Data(), testCase.DataLength)
			assert.EqualValues(testCase.Duration, preview.Duration())
			width, height := preview.Resolution()
			assert.Equal(testCase.Width, width)
			assert.Equal(testCase.Height, height)

			hash := sha256.Sum256(preview.Data())
			assert.Equal(testCase.Hash, hex.EncodeToString(hash[:]))
		})
	}
}
