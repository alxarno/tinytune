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
			DataLength: 34428,
			Duration:   time.Second * 5,
			Width:      1280,
			Height:     720,
			Hash:       "df04235b400d9e0d623681102ca40c7424ca5efff0014e4e1b74d0452c9030fb",
		},
		{
			Name:       ".flv",
			SourcePath: "../../test/video/sample_960x400_ocean_with_audio.flv",
			DataLength: 12534,
			Duration:   time.Second * 46,
			Width:      960,
			Height:     400,
			Hash:       "27374409e6121e4462ae96bc36ed9dc116913da3d45a297ccb6e92b9109a9be4",
		},
		{
			Name:       "short",
			SourcePath: "../../test/short.mp4",
			DataLength: 34500,
			Duration:   time.Second * 3,
			Width:      1280,
			Height:     720,
			Hash:       "60e585b7a0ad868d2c1b3237b6d163ea58faac560e1684604a7d8cca99a165b7",
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
